package controller

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/raikh/calc_micro_final/model"
)

type Task struct {
	ID            string
	ExpressionID  string
	Arg1          float64
	Arg2          float64
	Operation     string
	OperationTime int64    `json:"operation_time"`
	Dependencies  []string `json:"-"`
	Result        float64  `json:"-"`
	Completed     bool     `json:"-"`
	IsProcessing  bool     `json:"-"`
}

type Expression struct {
	ID     string
	Expr   string
	Status string
	Result float64
}

type ExpressionRequest struct {
	Expression string `json:"expression"`
}

func generateID() (uuid string) {

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return
}

func parseExpression(expr string, expressionID string, delayDict map[string]int64) []*model.Task {
	postfix := infixToPostfix(expr)
	stack := []*model.Task{}
	tasks := []*model.Task{}

	for _, token := range postfix {
		switch token {
		case "+", "-", "*", "/":
			arg2Task := stack[len(stack)-1]
			arg1Task := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			deps := []string{}
			if !arg1Task.Completed {
				deps = append(deps, arg1Task.Id)
			}
			if !arg2Task.Completed {
				deps = append(deps, arg2Task.Id)
			}

			task := &model.Task{
				Id:            generateID(),
				ExpressionId:  expressionID,
				Arg1:          arg1Task.Result,
				Arg2:          arg2Task.Result,
				Operation:     token,
				OperationTime: delayDict[token],
				Dependencies:  deps,
			}
			tasks = append(tasks, task)

			stack = append(stack, task)
		default:
			num, _ := strconv.ParseFloat(token, 64)

			task := &model.Task{
				Id:            generateID(),
				ExpressionId:  expressionID,
				Arg1:          &num,
				Arg2:          nil,
				Operation:     "",
				OperationTime: 0,
				Dependencies:  []string{},
				Result:        &num,
				Completed:     true,
			}
			stack = append(stack, task)
		}
	}

	return tasks
}

func infixToPostfix(expr string) []string {
	var output []string
	var stack []string

	precedence := map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
	}

	tokens := tokenize(expr)
	for _, token := range tokens {
		switch token {
		case "+", "-", "*", "/":
			for len(stack) > 0 && precedence[stack[len(stack)-1]] >= precedence[token] {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		case "(":
			stack = append(stack, token)
		case ")":
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = stack[:len(stack)-1]
		default:
			output = append(output, token)
		}
	}

	for len(stack) > 0 {
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output
}

func tokenize(expr string) []string {
	var tokens []string
	var currentToken string

	for _, char := range expr {
		if char == ' ' {
			continue
		}
		if char == '+' || char == '-' || char == '*' || char == '/' || char == '(' || char == ')' {
			if currentToken != "" {
				tokens = append(tokens, currentToken)
				currentToken = ""
			}
			tokens = append(tokens, string(char))
		} else {
			currentToken += string(char)
		}
	}

	if currentToken != "" {
		tokens = append(tokens, currentToken)
	}

	return tokens
}

func HandleGetExpressions() echo.HandlerFunc {
	return func(c echo.Context) error {

		user, ok := c.Get("user").(model.User)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		expressions, _ := model.GetExpressionsByUserId(user.Id)

		return c.JSON(http.StatusOK, map[string]interface{}{"expressions": expressions})
	}
}

func HandleGetExpressionsById() echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		user, ok := c.Get("user").(model.User)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		expressions, err := model.GetExpressionByIdForUser(id, user.Id)

		if err != nil {
			return c.JSON(http.StatusNotFound, err.Error())
		}

		return c.JSON(http.StatusOK, expressions)
	}
}

func HandleCalculate(delayDict map[string]int64) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(ExpressionRequest)
		if err := c.Bind(req); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		req.Expression = strings.TrimSpace(req.Expression)
		if req.Expression == "" {
			return c.JSON(http.StatusUnprocessableEntity, "Invalid request body")
		}

		user, ok := c.Get("user").(model.User)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		tx, err := model.BeginTx()
		if err != nil {
			return err
		}

		defer func() {
			if p := recover(); p != nil {
				tx.Rollback()
				panic(p)
			} else if err != nil {
				tx.Rollback()
			}
		}()

		id := generateID()
		now := time.Now()
		expr := &model.Expression{
			Id:         id,
			UserId:     user.Id,
			Expression: req.Expression,
			Status:     "pending",
			Result:     nil,
			CreatedAt:  &now,
			UpdatedAt:  &now,
		}
		err = expr.InsertTx(tx)
		if err != nil {
			return err
		}

		tasksForExpr := parseExpression(req.Expression, id, delayDict)
		for _, task := range tasksForExpr {
			task.CreatedAt = &now
			task.UpdatedAt = &now
			err = task.InsertTx(tx)
			if err != nil {
				return err
			}
		}

		tx.Commit()

		return c.JSON(http.StatusCreated, map[string]string{"id": id})
	}
}
