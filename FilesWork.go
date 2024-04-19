package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"gopkg.in/ini.v1"
)

// writeToFile записывает данные data []byte в файл с именем fileName
func writeToFile(fileName string, data []byte) {
	file, err := os.Create(fileName) // создаем файл
	writer := bufio.NewWriter(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	if _, err := writer.Write(data); err != nil { // записываем данные в файл
		fmt.Println(err)
	}
	writer.Flush() // сбрасываем данные из буфера в файл
	fmt.Println("Файл записан", fileName)
}

// readSettingFromINI считывает из указанного INI-файла URL сервера Wialon API и токен безопасности
func readSettingFromINI(fileName string) (URL, Token string) {
	cfg, err := ini.Load(fileName)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	// Классическое чтение значений, раздел (Section) по умолчанию может быть представлен пустой строкой.
	URL = cfg.Section("Wialon").Key("URL").String()
	Token = cfg.Section("Wialon").Key("Token").String()
	return
}

// getNextRune инкрементирует букву английского алфавита A->B, B->C, ... Z-A
func getNextLetter(letter string) string {

	letters := []rune(letter)
	if len(letters) == 0 {
		return letter
	}

	prefix := string(letters[:len(letters)-1])
	if letters[len(letters)-1] == 'Z' {
		if len(letters) == 1 {
			prefix = "A"
		} else {
			prefix = getNextLetter(string(letters[:len(letters)-1]))
		}
	}
	ch := letters[len(letters)-1]

	return prefix + string((ch+1-'A')%('Z'-'A'+1)+'A')
}

// Excel-файл пакета excelize
var excelFile = excelize.NewFile()

// exportJsonTableToExcel экспортирует данные из таблицы джейсон
func exportJsonTableToExcel(table jsonTable, sheet, fileName string) {
	excelFile.DeleteSheet("Sheet1")
	excelFile.NewSheet(sheet)
	col := "A"
	row := 0

	for _, element := range table {
		for _, group := range element.Groups {
			for _, data := range group.Rows {
				exportToCell(sheet, col, strconv.Itoa(row), data)
				col = getNextLetter(col)
			}
			col = "A"
			row++
		}
	}

	excelFile.SaveAs(fileName + ".xlsx")
}

// exportToCell экспортирует любые данные в указанную ячейку Excel
func exportToCell(sheet, col, row string, val any) {
	cell := col + row
	//println(cell)
	switch val.(type) {
	case string:
		cellWidth := 2.0
		if text, ok := val.(string); ok {
			switch {
			case len([]rune(text)) < 6 && strings.Index(text, ".") > 0:
				cellWidth = 5
			case len([]rune(text)) < 12:
				cellWidth = 14
			case len([]rune(text)) < 20:
				cellWidth = 20
			default:
				cellWidth = 30
			}
		}
		excelFile.SetColWidth(sheet, col, col, cellWidth)
		excelFile.SetCellValue(sheet, cell, val)
	case float64:
		excelFile.SetColWidth(sheet, col, col, 10)
		excelFile.SetCellValue(sheet, cell, val)
	case map[string]interface{}:
		if timeCoords, ok := val.(map[string]interface{}); ok {
			if floatTime, ok := timeCoords["v"].(float64); ok {
				unixTime := time.Unix(int64(floatTime), 0)
				//fmt.Printf("%20v |", unixTime.Format("02.01.2006 15:04:05"))
				excelFile.SetColWidth(sheet, col, col, 20)
				excelFile.SetCellValue(sheet, cell, unixTime)
			} else if text, ok := timeCoords["t"]; ok {
				//fmt.Printf("%-55s |", text)
				excelFile.SetColWidth(sheet, col, col, 60)
				excelFile.SetCellValue(sheet, cell, text)
			}
		}
	default:
		//fmt.Printf("ХЗ: %-10v | ", val)
		excelFile.SetColWidth(sheet, col, col, 10)
		excelFile.SetCellValue(sheet, cell, val)
	}
}
