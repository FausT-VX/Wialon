package main

import (
	"fmt"
	"strings"
	"time"
)

// PrintJTable печатает таблицу по данным структуры jsonTable
func PrintJTable(table jsonTable) {
	for _, element := range table {
		for _, group := range element.Groups {
			for _, row := range group.Rows {
				PrintRow(row)
			}
			fmt.Println()
		}
	}
}

// PrintRow печатает одну строку данных таблицы из JSON в отформатированном виде.
func PrintRow(val any) {
	switch val.(type) {
	case string:
		format := ""
		if text, ok := val.(string); ok {
			switch {
			case len([]rune(text)) < 6 && strings.Index(text, ".") > 0:
				format = "%-5s"
			case len([]rune(text)) < 12:
				format = "%-11s"
			case len([]rune(text)) < 20:
				format = "%-20s"
			default:
				format = "%-30s"
			}
		}
		fmt.Printf(format+" | ", val)
	case float64:
		fmt.Printf("%-10.2f | ", val)
	case map[string]interface{}:
		if timeCoords, ok := val.(map[string]interface{}); ok {
			if floatTime, ok := timeCoords["v"].(float64); ok {
				unixTime := time.Unix(int64(floatTime), 0)
				fmt.Printf("%20v |", unixTime.Format("02.01.2006 15:04:05"))
			} else if text, ok := timeCoords["t"]; ok {
				fmt.Printf("%-55s |", text)
			}
		}
	default:
		fmt.Printf("ХЗ: %-10v | ", val)
	}
}

// PrintAny печатает любой JSON (any, map[string]interface{})
func PrintAny(val any) {
	switch val.(type) {
	case string:
		fmt.Printf("%-15s | ", val)
	case float64:
		fmt.Printf("%-10.2f | ", val)
	case map[string]interface{}:
		fmt.Printf("\n%50s", "")
		if element, ok := val.(map[string]interface{}); ok {
			for _, v := range element {
				PrintAny(v)
			}
		}
	default:
		fmt.Printf("ХЗ: %-10v | ", val)
	}
}
