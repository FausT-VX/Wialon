package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	//"github.com/sanity-io/litter"
)

// URL сервера Wialon API
var WialonURL = ""

// Токен безопасности для первичной авторизации
var Token = ""

var y, m, d = time.Now().Date()
var loc = time.Now().Location()
var dateBeg = time.Date(y, m, d, 8, 00, 00, 0, loc).Add(-time.Hour * 24).Unix() // Дата начала отчета
var dateEnd = time.Date(y, m, d, 9, 59, 59, 0, loc).Add(-time.Hour * 24).Unix() // Дата окончания отчета

// Структура данных получаемого JSON
type jsonTable []struct {
	// N   int   `json:"n"`
	// I1  int   `json:"i1"`
	// I2  int   `json:"i2"`
	// T1  int   `json:"t1"`
	// T2  int   `json:"t2"`
	// D   int   `json:"d"`
	// Mrk int   `json:"mrk"`
	// C   []any `json:"c"`
	Groups []struct { //r
		// N   int   `json:"n"`
		// I1  int   `json:"i1"`
		// I2  int   `json:"i2"`
		// T1  int   `json:"t1"`
		// T2  int   `json:"t2"`
		// D   int   `json:"d"`
		// Mrk int   `json:"mrk"`
		Rows []any `json:"c"`
	} `json:"r"`
}

func main() {
	WialonURL, Token = readSettingFromINI("Wialon.ini")
	sid := WialonAPI_GetSID()
	println(sid)
	WialonAPI_ExecReport(sid)
	PrintJTable(WialonAPI_GetTable(sid, 0))
	PrintJTable(WialonAPI_GetTable(sid, 1))
	PrintJTable(WialonAPI_GetTable(sid, 2))
}

// WialonAPI_GetSID используя Wialon-API подключается к серверу Wialon используя токен-безопасности
// и возвращает SID используемый для дальнейшей авторизации в сессии
func WialonAPI_GetSID() string {
	// определяем URL для обращения и его параметры
	params := url.Values{}
	params.Add("Content-Type", "application/x-www-form-urlencoded")
	params.Add("svc", "token/login")

	m := map[string]string{"token": Token}
	mJson, err := json.Marshal(m)
	if err != nil {
		fmt.Println("ошибка преобразования параметра params:", err)
		return ""
	}
	params.Add("params", string(mJson))

	// преобразуем URL-строку в объект и заполняем параметры
	urlInstance, err := url.Parse(WialonURL)
	if err != nil {
		fmt.Println("ошибка парсинга URL", err)
		return ""
	}
	urlInstance.RawQuery = params.Encode()

	// выполняем запрос и обрабатываем ответ
	response, err := http.Get(urlInstance.String())
	if err != nil {
		fmt.Println("ошибка выполнения запроса:", err)
		return ""
	}
	//fmt.Println(response.Status)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	//fmt.Println(string(body))

	var jResp map[string]interface{}
	err = json.Unmarshal(body, &jResp)
	if err != nil {
		fmt.Println("ошибка чтения JSON-данных:", err)
	}

	return fmt.Sprint(jResp["eid"])
}

// WialonAPI_ExecReport запускает на выполнение отчет на сервере Wialon по указанному идентификатору сессии (SID)
func WialonAPI_ExecReport(sid string) {
	jParams := map[string]interface{}{
		"reportResourceId":  14204627,
		"reportTemplateId":  14,
		"reportTemplate":    nil,
		"reportObjectId":    14204751,
		"reportObjectSecId": 0,
		"interval": map[string]interface{}{
			"flags": 16777216,
			"from":  dateBeg,
			"to":    dateEnd,
		},
	}

	WialonAPI_SendRequest(sid, "report/exec_report", jParams)
}

// WialonAPI_GetTable получает с сервера Wialon таблицу с результатами выполнения отчета
// по указанному идентификатору сессии (SID) и индексу таблицы (tableIndex)
func WialonAPI_GetTable(sid string, tableIndex int) jsonTable {

	jParams := map[string]interface{}{
		"tableIndex": tableIndex,
		"config": map[string]interface{}{
			"type": "range",
			"data": map[string]interface{}{
				"from": 0, "to": 999999, "level": 2,
			},
		},
	}

	response := WialonAPI_SendRequest(sid, "report/select_result_rows", jParams)
	buffer := bytes.NewBuffer([]byte{})
	if err := json.Indent(buffer, response, "", "  "); err != nil {
		fmt.Println(err)
		return nil
	}
	writeToFile(fmt.Sprint("RespTable", tableIndex, ".json"), buffer.Bytes())

	var jResp jsonTable
	if err := json.Unmarshal(response, &jResp); err != nil {
		fmt.Println("(WialonAPI_GetTable) ошибка чтения JSON-данных:", err)
	}
	return jResp
}

// WialonAPI_SendRequest отправляет запрос на сервер Wialon по указанному идентификатору сессии (SID),
// вызываемому сервису svc c параметрами jParams и возвращает тело ответа (response.Body)
func WialonAPI_SendRequest(sid, svc string, jParams map[string]interface{}) (respBody []byte) {
	// сериализация параметров
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	// Запись JSON-данных в буфер
	if err := encoder.Encode(jParams); err != nil {
		fmt.Println("Ошибка при записи JSON-данных:", err)
		return
	}
	//fmt.Println(buf.String())
	// определяем URL для обращения и его параметры
	params := url.Values{}
	params.Add("svc", svc)
	params.Add("sid", sid)
	params.Add("params", buf.String())

	// преобразуем URL-строку в объект и заполняем параметры
	urlInstance, err := url.Parse(WialonURL)
	if err != nil {
		fmt.Println("ошибка парсинга URL", err)
		return
	}
	//Способ 1 ==============================================================================
	urlInstance.RawQuery = params.Encode()

	// выполняем запрос и обрабатываем ответ
	response, err := http.Get(urlInstance.String())
	if err != nil {
		fmt.Println("ошибка выполнения запроса:", err)
		return
	}
	fmt.Println(response.Status)

	respBody, err = io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Println(string(respBody))
	return respBody

	//Способ 2 ==============================================================================

	/*data := []byte(params.Encode())
	// Создаем новый запрос
	req, err := http.NewRequest("POST", urlInstance.String(), bytes.NewReader(data))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Устанавливаем заголовок с типом данных в теле запроса
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Выводим ответ от сервера
	fmt.Println(resp.Status)
	body2, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	*/
}
