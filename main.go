package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const sourceUrl = "https://pub.fsa.gov.ru"
const username = "anonymous"
const password = "hrgesf7HDR67Bd"

const sizeDeclarations = 20000

var bearerToken string
var initRequest *gorequest.SuperAgent

func main() {

	initBearerToken()
	runCollectDeclarationIds()

}

func initBearerToken() {
	defer exception("getBearerToken failed")
	bearerToken = getBearerToken()
	fmt.Println(bearerToken)
}

func runCollectDeclarationIds() {
	defer exception("collectDeclarationIds failed")
	collectDeclarationIds()
	fmt.Println("Done Collect Declaration Ids!")
}

func collectDeclarationIds() {
	currentPage := 1
	totalPages := 0
	totalItems := 0

	process := true

	for process {
		declarations := getDeclarations(currentPage, sizeDeclarations)

		if totalItems == 0 {
			totalItems = int(declarations["total"].(float64))
			totalPages = totalItems / sizeDeclarations
		}

		declarationIds := getDeclarationIds(declarations)
		declarationIdsLen := len(declarationIds)

		if currentPage > totalPages {
			process = false
		}

		fmt.Println(totalPages, currentPage, declarationIdsLen)

		currentPage++
	}
}

func getDeclarationIds(declarations map[string]interface{}) []int {
	items := declarations["items"].([]interface{})

	ItemsLen := len(items)
	declarationIds := make([]int, ItemsLen)

	for i := 0; ItemsLen > i; i++ {
		item := items[i].(map[string]interface{})
		declarationIds[i] = int(item["id"].(float64))
	}

	return declarationIds
}

func getDeclarations(page, size int) map[string]interface{} {
	response, _, _ := makeRequest().
		Post(sourceUrl+"/api/v1/rds/common/declarations/get").
		AppendHeader("Authorization", bearerToken).
		Send(fmt.Sprintf(getRequestBody(), size, page)).
		End()

	return readCloserToJson(response.Body)
}

func readCloserToJson(data io.ReadCloser) map[string]interface{} {
	dataInBytes, err := ioutil.ReadAll(data)
	if err != nil {
		panic("ReadAll: " + err.Error() + "; " + string(dataInBytes))
	}

	var jsonMap map[string]interface{}

	err = json.Unmarshal(dataInBytes, &jsonMap)
	if err != nil {
		panic("Unmarshal: " + err.Error() + "; " + string(dataInBytes))
	}

	return jsonMap
}

func getRequestBody() string {
	return `{"size":%v,"page":%v,"filter":{"status":[],"idDeclType":[],"idCertObjectType":[],"idProductType":[],"idGroupRU":[],"idGroupEEU":[],"idTechReg":[],"idApplicantType":[],"regDate":{"minDate":null,"maxDate":null},"endDate":{"minDate":null,"maxDate":null},"columnsSearch":[{"name":"number","search":"","type":0,"translated":false}],"idProductOrigin":[],"idProductEEU":[],"idProductRU":[],"idDeclScheme":[],"awaitForApprove":null,"editApp":null,"violationSendDate":null},"columnsSort":[{"column":"declDate","sort":"DESC"}]}`
}

func getBearerToken() string {
	response, _, errs := makeRequest().
		Post(sourceUrl + "/login").
		Send(`{"username":"` + username + `","password":"` + password + `"}`).
		End()

	if errs != nil {
		panic(errs)
	}

	return response.Header.Get("Authorization")
}

func makeRequest() *gorequest.SuperAgent {
	if initRequest == nil {
		initRequest = gorequest.New().
			Retry(5, 3*time.Second, http.StatusBadGateway, http.StatusBadRequest, http.StatusInternalServerError).
			TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
			AppendHeader("Content-Type", "application/json")
	}

	return initRequest.Clone()
}

func exception(message string) {
	if err := recover(); err != nil {
		fmt.Println(message+":", err)
	}
}
