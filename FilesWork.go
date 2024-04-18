package main

import (
	"bufio"
	"fmt"
	"os"

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
