package gorequests

import "fmt"

type (
	QueryParam struct {
		Key       string `json:"key"`
		ValueType string `json:"value_type"`
		Value     string `json:"value"`
	}
)

func (ast *ApiStore) FormQueryParams(req *Request) (queryParamsStr string, err error) {
	for idx, queryParam := range req.QueryParams {
		var (
			key, value string
		)
		if idx == 0 {
			queryParamsStr += "?"
		} else {
			queryParamsStr += "&"
		}
		if key, value, err = queryParam.GetKeyValue(ast); err != nil {
			return
		}
		queryParamsStr += fmt.Sprintf("%s=%s", key, value)

	}
	return
}

func (qp *QueryParam) GetKeyValue(ast *ApiStore) (key, value string, err error) {
	switch qp.ValueType {
	case "static":
		key, value, err = qp.GetStaticKeyValue(ast)
		break
	default:
		key, value, err = qp.GetStaticKeyValue(ast)
		break
	}
	return
}

func (qp *QueryParam) GetStaticKeyValue(ast *ApiStore) (key, value string, err error) {
	key = qp.Key
	value = qp.Value
	return
}
