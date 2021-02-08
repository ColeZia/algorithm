package arithmetic

import (
	"bytes"
	"container/list"
	"encoding/json"
	"errors"
	"strconv"
	"unicode"
)

const ExpressionError = "expression error"

type stack struct {
	listStack []interface{}
}

func (s *stack) pop() string {
	value := s.listStack[len(s.listStack)-1]
	s.listStack = s.listStack[:len(s.listStack)-1]
	return value.(string)
}

func (s *stack) push(item interface{}) {
	s.listStack = append(s.listStack, item)
}

func (s *stack) getPop() string {
	return s.listStack[len(s.listStack)-1].(string)
}

var ParenthesesMap = map[int32]bool{
	'(': true,
	')': true,
}

var OperatorSymbol = map[int32]bool{
	'+': true,
	'-': true,
	'*': true,
	'/': true,
	'(': true,
	')': true,
}

var OperatorStrSymbol = map[string]bool{
	"+": true,
	"-": true,
	"*": true,
	"/": true,
	"(": true,
	")": true,
}

type ListMeta struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

func Expression(expression string, list string) string {
	//处理并验证四则运算
	items, err := parseExpression(expression)
	if err != nil {
		return err.Error()
	}
	items, err = replaceKeyToValue(list, items)
	if err != nil {
		return ExpressionError
	}
	//数栈
	numStack := &stack{
		listStack: []interface{}{},
	}
	//运算符栈
	symbolStack := &stack{
		listStack: []interface{}{},
	}
	for _, item := range items {
		if _, ok := OperatorStrSymbol[item]; ok {
			//运算符栈
			if len(symbolStack.listStack) == 0 {
				symbolStack.push(item)
			} else {
				top := symbolStack.getPop()
				//优先级大运算符的入栈
				if isPushStack(item, top) {
					symbolStack.push(item)
				} else {
					right := numStack.pop()
					symbol := symbolStack.pop()
					left := numStack.pop()
					res := calculateExpression(left, right, symbol)
					numStack.push(res)
					symbolStack.push(item)
				}
			}
		}
		if isNumberItem(item) {
			//数栈
			numStack.push(item)
		}
		if item == ")" {
			right := numStack.pop()
			symbolStack.pop()
			symbol := symbolStack.pop()
			left := numStack.pop()
			res := calculateExpression(left, right, symbol)
			numStack.push(res)
			if symbolStack.getPop() == "(" {
				symbolStack.pop()
				if symbolStack.getPop() == "-" {
					symbolStack.pop()
					symbolStack.push("+")
					result, err := negativeToPositive(numStack.pop())
					if err != nil {
						panic(err)
					}
					numStack.push(result)
				}
			}
		}
	}
	if len(symbolStack.listStack) >= len(numStack.listStack) {
		return ExpressionError
	}
	for len(symbolStack.listStack) != 0 {
		right := numStack.pop()
		symbol := symbolStack.pop()
		left := numStack.pop()
		res := calculateExpression(left, right, symbol)
		numStack.push(res)
	}
	result := numStack.pop()
	_, err = strconv.ParseFloat(result, 64)
	if err != nil {
		return ExpressionError
	}
	return result
}

//处理并验证四则运算
func parseExpression(expression string) ([]string, error) {
	var runes []rune
	var stack list.List
	for _, item := range expression {
		if _, ok := ParenthesesMap[item]; ok {
			if stack.Len() == 0 {
				stack.PushBack(item)
			} else if isSame(stack.Back().Value.(int32), item) {
				stack.Remove(stack.Back())
			} else {
				stack.PushBack(item)
			}
		}
		runes = append(runes, item)
	}
	if stack.Len() != 0 {
		return nil, errors.New(ExpressionError)
	}
	var items []string
	specialSymbol := ""
	for i := 0; i < len(runes); {
		var position int
		var value string
		if isNumber(runes[i]) {
			value, position = readCharters(runes, i, isNumber)
			if specialSymbol != "" {
				value = specialSymbol + value
				specialSymbol = ""
			}
		}
		if unicode.IsLetter(runes[i]) {
			value, position = readCharters(runes, i, unicode.IsLetter)
		}
		if unicode.IsSpace(runes[i]) {
			i++
			continue
		}
		if _, ok := OperatorSymbol[runes[i]]; ok {
			value = string(runes[i])
			i += 1
			position = i
			if i < len(runes) && isNextNumber(runes[i]) {
				if value == "-" {
					specialSymbol = value
					continue
				}
			}
		}
		items = append(items, value)
		i = position
	}
	return items, nil
}

//拆分元数据
func readCharters(runes []rune, position int, condition func(rune2 rune) bool) (string, int) {
	var itemBuffer bytes.Buffer
	for j := position; j < len(runes); {
		character := runes[j]
		if unicode.IsSpace(character) {
			position = j
			break
		}
		if condition(character) {
			itemBuffer.WriteString(string(character))
		} else {
			position = j
			break
		}
		j += 1
		position = j
	}
	return itemBuffer.String(), position
}

func replaceKeyToValue(list string, items []string) ([]string, error) {
	values := []ListMeta{}
	valueMap := make(map[string]float64)
	err := json.Unmarshal([]byte(list), &values)
	if err != nil {
		return nil, errors.New(ExpressionError)
	}
	for _, value := range values {
		valueMap[value.Key] = value.Value
	}
	//替换变量
	for i := 0; i < len(items); i++ {
		if value, ok := valueMap[items[i]]; ok {
			items[i] = strconv.FormatFloat(value, 'f', -1, 64)
		}
	}
	return items, nil
}

func calculateExpression(left, right interface{}, symbol string) string {
	var res float64
	switch symbol {
	case "+":
		res = add(left, right)
	case "-":
		res = sub(left, right)
	case "*":
		res = multiplication(left, right)
	case "/":
		res = division(left, right)
	}
	return strconv.FormatFloat(res, 'f', -1, 64)
}

func negativeToPositive(item string) (string, error) {
	res, err := strconv.ParseFloat(item, 64)
	if err != nil {
		return "", err
	}
	res = -res
	result := strconv.FormatFloat(res, 'f', -1, 64)
	return result, nil
}

//当前优先级大于top 继续入栈
func isPushStack(cur, top string) bool {
	pCur, pTop := 0, 0
	if cur == "(" {
		pCur = 3
	} else if cur == "+" || cur == "-" {
		pCur = 1
	} else {
		pCur = 2
	}
	if top == "(" {
		pTop = 0
	} else if top == "+" || top == "-" {
		pTop = 1
	} else if top == "*" || top == "/" {
		pTop = 2
	} else {
		pTop = 3
	}
	return pCur >= pTop
}

func isNumber(item int32) bool {
	return unicode.IsDigit(item) || item == '.'
}

func isNumberItem(item string) bool {
	_, err := strconv.Atoi(item)
	if err != nil {
		return false
	}
	return true
}

//验证括号完整性
func isSame(s1, s2 int32) bool {
	return s1 == '(' && s2 == ')'
}

func isNextNumber(item int32) bool {
	return unicode.IsDigit(item)
}

func add(left, right interface{}) float64 {
	l, err := strconv.ParseFloat(left.(string), 64)
	if err != nil {
		return 0
	}
	r, err := strconv.ParseFloat(right.(string), 64)
	if err != nil {
		return 0
	}
	return l + r
}

func sub(left, right interface{}) float64 {
	l, err := strconv.ParseFloat(left.(string), 64)
	if err != nil {
		return 0
	}
	r, err := strconv.ParseFloat(right.(string), 64)
	if err != nil {
		return 0
	}
	return l - r
}

func multiplication(left, right interface{}) float64 {
	l, err := strconv.ParseFloat(left.(string), 64)
	if err != nil {
		return 0
	}
	r, err := strconv.ParseFloat(right.(string), 64)
	if err != nil {
		return 0
	}
	return l * r
}
func division(left, right interface{}) float64 {
	l, err := strconv.ParseFloat(left.(string), 64)
	if err != nil {
		return 0
	}
	r, err := strconv.ParseFloat(right.(string), 64)
	if err != nil {
		return 0
	}
	return l / r
}
