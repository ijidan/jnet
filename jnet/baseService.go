package jnet

//基本服务
type BaseService struct {
	useProxy      bool
	statusSymbol  string
	messageSymbol string
}

//设置可变变量
func (bs *BaseService) SetCusConfig(statusSymbol string, messageSymbol string) {
	bs.statusSymbol = statusSymbol
	bs.messageSymbol = messageSymbol
}

//设置是否使用代理
func (bs *BaseService) SetUseProxy(useProxy bool) {
	bs.useProxy = useProxy
}

//GET请求
func (bs *BaseService) SendGetRequest(host string, params map[string]interface{}) Response {
	return bs.sendRequest(MethodGet, host, params)
}

//POST请求
func (bs *BaseService) SendPostRequest(host string, params map[string]interface{}) Response {
	return bs.sendRequest(MethodPost, host, params)
}

//PUT请求
func (bs *BaseService) SendPutRequest(host string, params map[string]interface{}) Response {
	return bs.sendRequest(MethodPut, host, params)
}

//DELETE请求
func (bs *BaseService) SendDeleteRequest(host string, params map[string]interface{}) Response {
	return bs.sendRequest(MethodDelete, host, params)
}

//发送请求
func (bs *BaseService) sendRequest(method string, host string, param map[string]interface{}) Response {
	request := NewRequest()
	var useProxy bool
	if bs.useProxy == true {
		useProxy = true
	} else {
		useProxy = false
	}
	request.SetUseProxy(useProxy)
	request.SetMethod(method)
	request.SetUUID("")
	request.SetUrl(host)
	request.SetParams(param)
	rsp := request.send()
	return rsp
}
