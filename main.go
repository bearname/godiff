package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
)

func main() {
	var leftFileName string
	var rightFileName string
	var outputFilename string
	flag.StringVar(&leftFileName, "left", "left.txt", "file path")
	flag.StringVar(&rightFileName, "right", "right.txt", "file path")
	flag.StringVar(&outputFilename, "output", "", "file path")
	flag.Parse()
	leftTokens, err := FileToArray(leftFileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	rightTokens, err := FileToArray(rightFileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	first2Second, second2First := getDiff(leftTokens, rightTokens)
	var outputFile *os.File
	if len(outputFilename) != 0 {
		outputFile, err = os.Create(outputFilename)
		if err != nil {
			fmt.Println(err)
			return
		}

		result := getHeaderHtml() + "<div style=\"display: flex;\">" +
			" <div style=\"width: 50%; overflow-y: scroll; \">" +
			buildTableHtml(first2Second, "Left diff right") +
			"</div> <div style=\"width: 50%; overflow-y: scroll; \">" +
			buildTableHtml(second2First, "Right diff left") +
			"</div></div>" +
			getFooterHtml()
		_, err = outputFile.WriteString(result)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Success write to " + outputFilename)

		}
	} else {
		drawOnStdOut(first2Second, second2First)
	}
}

func buildTableHtml(first2Second []TokenItem, msg string) string {
	result := getTableHeaderHtml(msg)
	for _, item := range first2Second {
		result += getRowHtml(item)
	}
	result += getTableFooterHtml()
	return result
}

func getHeaderHtml() string {
	return `<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Document</title>
   <style>
        tr,
        td,
        tbody,
        table {
           border: 0 none;
           border-spacing: 0;
        }
    </style>
</head>
<body>`
}

func getTableFooterHtml() string {
	return `</tbody>
</table>`
}
func getFooterHtml() string {
	return `</body>
</html>`
}

func getRowHtml(token TokenItem) string {
	color := getValueColorHtml(token)

	action := getAction(token)

	return `<tr style="background-color:` + color + `;">
        <td>` + strconv.Itoa(token.CurrPosition) + ` ` + action + ` </td>
        <td>` + token.Token + `</td>
    </tr>`
}

func getAction(token TokenItem) string {
	action := "="

	switch token.Status {
	case Added:
		action = "+"
	case NotChanged:
		action = "="
	case Removed:
		action = "-"
	}
	return action
}

func getValueColorHtml(token TokenItem) string {
	var color = "white"

	switch token.Status {
	case Added:
		color = "green"
	case NotChanged:
		color = "white"
	case Removed:
		color = "red"
	}
	return color
}

func getTableHeaderHtml(msg string) string {
	return `
<h3>` + msg + `</h3>
<table>
    <tr>
    <th>line</th>
    <th>content</th>
    </tr>
    <tbody>`
}

func drawOnStdOut(first2Second []TokenItem, second2First []TokenItem) {
	fmt.Println(" First diff second")
	drawResult(first2Second)
	fmt.Println("\n\n\n\n\n\n\n\n Second diff first")
	drawResult(second2First)
}
func FileToArray(filename string) ([]string, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}

	return txtlines, nil
}

var (
	Red   = Color("\033[1;31m%s\033[0m")
	Green = Color("\033[1;32m%s\033[0m")
	White = Color("\033[1;37m%s\033[0m")
)

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

func getDiff(leftTokens []string, rightTokens []string) ([]TokenItem, []TokenItem) {
	matrix := buildMatrix(rightTokens, leftTokens)

	first2Second, second2First := restoreLineSequence(leftTokens, rightTokens, matrix)
	first2Second = Reverse(first2Second)
	second2First = Reverse(second2First)
	return first2Second, second2First
}

func buildMatrix(secondTokens []string, firstTokens []string) [][]int {
	firstLen := len(firstTokens)
	var matrix = make([][]int, firstLen)
	secondLen := len(secondTokens)
	for i := range matrix {
		matrix[i] = make([]int, secondLen)
	}
	for i := 0; i < firstLen; i++ {
		for j := 0; j < secondLen; j++ {
			if i == 0 || j == 0 {
				matrix[i][j] = 0
			} else if firstTokens[i-1] == secondTokens[j-1] {
				matrix[i][j] = 1 + matrix[i-1][j-1]
			} else {
				max := int(math.Max(float64(matrix[i-1][j]), float64(matrix[i][j-1])))
				matrix[i][j] = max
			}
		}
	}
	return matrix
}

func restoreLineSequence(firstTokens []string, secondTokens []string, matrix [][]int) ([]TokenItem, []TokenItem) {
	firstLen := len(firstTokens)
	secondLen := len(secondTokens)
	var first2Second []TokenItem
	var second2First []TokenItem
	var tokenItemLhs TokenItem
	var tokenItemRhs TokenItem

	var token string
	i := firstLen - 1
	j := secondLen - 1
	for i != 0 || j != 0 {
		tokenItemLhs = TokenItem{
			Token:        firstTokens[i-1],
			CurrPosition: i,
			Status:       NotChanged,
		}
		tokenItemRhs = TokenItem{
			Token:        firstTokens[i-1],
			CurrPosition: j,
			Status:       NotChanged,
		}
		if i == 0 {
			tokenItemLhs.setStatus(Added)
			tokenItemRhs.setStatus(Removed)
			token = secondTokens[j-1]
			tokenItemLhs.setPosition(j-1, j)
			tokenItemRhs.setPosition(j, j-1)
			j--
		} else if j == 0 {
			tokenItemLhs.setStatus(Removed)
			tokenItemRhs.setStatus(Added)
			token = firstTokens[i-1]
			tokenItemLhs.setPosition(i-1, i)
			tokenItemRhs.setPosition(i, i-1)
			i--
		} else if firstTokens[i-1] == secondTokens[j-1] {
			token = firstTokens[i-1]
			i--
			j--
			tokenItemLhs.setPosition(i, i)
			tokenItemRhs.setPosition(i, i)
		} else if matrix[i-1][j] <= matrix[i][j-1] {
			tokenItemLhs.setStatus(Added)
			tokenItemRhs.setStatus(Removed)
			token = secondTokens[j-1]
			tokenItemLhs.setPosition(j-1, j)
			tokenItemLhs.setPosition(j, j-1)
			j--
		} else {
			tokenItemLhs.setStatus(Removed)
			tokenItemRhs.setStatus(Added)
			token = firstTokens[i-1]
			tokenItemLhs.setPosition(i-1, i)
			tokenItemLhs.setPosition(i, i-1)
			i--
		}
		tokenItemLhs.Token = token
		tokenItemRhs.Token = token
		first2Second = append(first2Second, tokenItemLhs)
		second2First = append(second2First, tokenItemRhs)
	}
	return first2Second, second2First
}

func drawResult(tokenResult []TokenItem) {
	for _, item := range tokenResult {
		token := getConsoleOutput(item)
		fmt.Println(token)
	}
}

func getConsoleOutput(item TokenItem) string {
	token := strconv.Itoa(item.CurrPosition) + " "
	switch item.Status {
	case Added:
		token += "+ "
	case NotChanged:
		token += "= "
	case Removed:
		token += "- "
	}

	token += item.Token
	switch item.Status {
	case Added:
		token = Green(token)
	case NotChanged:
		token = White(token)
	case Removed:
		token = Red(token)
	}
	return token
}

func Reverse(input []TokenItem) []TokenItem {
	var output []TokenItem

	for i := len(input) - 1; i >= 0; i-- {
		output = append(output, input[i])
	}

	return output
}

type Status int

const (
	Removed Status = iota
	NotChanged
	Added
)

type TokenItem struct {
	Token        string
	CurrPosition int
	OldPosition  int
	Status       Status
}

func (t *TokenItem) setPosition(old, newPos int) {
	t.OldPosition = old
	t.CurrPosition = newPos
}

func (t *TokenItem) setStatus(status Status) {
	t.Status = status
}
