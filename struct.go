package main

type UpdateRequest struct {
	ZoneName string       `json:"zone_name"`
	Contents []ContentSet `json:"contents"`
}

type ContentSet struct {
	HostName string `json:"host_name"`
	Content  string `json:"content"`
}

type UpdateResponse struct {
	ZoneName string        `json:"zone_name"`
	Results  []ResponseSet `json:"results"`
}

type ResponseSet struct {
	Name      string `json:"name"`
	Content   string `json:"content"`
	Succeeded bool   `json:"succeeded"`
	Error     string `json:"error"`
}

type ErrorMessage struct {
	Code    string `json:"error"`
	Message string `json:"message"`
}

var (
	// E101: リクエストヘッダーにAuthorizationが存在しません
	errNotFoundAuthorizationHeader = ErrorMessage{
		Code:    "E101",
		Message: "リクエストヘッダーにAuthorizationが存在しません",
	}

	// E102: Authorizationヘッダーが不正です
	errInvalidAuthorizationHeader = ErrorMessage{
		Code:    "E102",
		Message: "Authorizationヘッダーが不正です",
	}

	// E103: リクエストされたデータが不正です
	errInvalidRequestedData = ErrorMessage{
		Code:    "E103",
		Message: "リクエストされたデータが不正です",
	}

	// E201: 指定されたゾーン情報が見つかりません
	errNotFoundZone = ErrorMessage{
		Code:    "E201",
		Message: "指定されたゾーン情報が見つかりません",
	}
)
