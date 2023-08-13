package main

import (
	"caiyun/sdkInit"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	cc_version = "1.0.0"
)

type Chaincode struct {
	Name string
	Path string
}

var (
	App  sdkInit.Application
	orgs []*sdkInit.OrgInfo
	info sdkInit.SdkEnvInfo
)

func main() {
	// init orgs information
	orgs = []*sdkInit.OrgInfo{
		{
			OrgAdminUser:  "Admin",
			OrgName:       "Org1",
			OrgMspId:      "Org1MSP",
			OrgUser:       "User1",
			OrgPeerNum:    2,
			OrgAnchorFile: "/root/go/src/caiyun/fixtures/channel-artifacts/Org1MSPanchors.tx",
		},
		{
			OrgAdminUser:  "Admin",
			OrgName:       "Org2",
			OrgMspId:      "Org2MSP", // 根据实际情况设置 Org2MSP 的值
			OrgUser:       "User1",
			OrgPeerNum:    2,                                                                                      // 根据实际情况设置 Org2 的 Peer 数量
			OrgAnchorFile: "/root/go/src/caiyun/fixtures/channel-artifacts/Org2MSPanchors.tx", // 根据实际情况设置 Org2MSP 的锚节点文件路径
		},
	}

	// init sdk env info
	chaincodes := []Chaincode{
		{"nft", "/root/go/src/caiyun/chaincode/nft/"},
		{"image", "/root/go/src/caiyun/chaincode/image/"},
		{"product", "/root/go/src/caiyun/chaincode/product/"},
	}
	info = sdkInit.SdkEnvInfo{
		ChannelID:        "mychannel",
		ChannelConfig:    "/root/go/src/caiyun/fixtures/channel-artifacts/channel.tx",
		Orgs:             orgs,
		OrdererAdminUser: "Admin",
		OrdererOrgName:   "OrdererOrg",
		OrdererEndpoint:  "orderer.example.com",
	}

	// sdk setup
	sdk, err := sdkInit.Setup("config.yaml", &info)
	if err != nil {
		fmt.Println(">> SDK setup error:", err)
		os.Exit(-1)
	}

	// create channel and join
	if err := sdkInit.CreateAndJoinChannel(&info); err != nil {
		fmt.Println(">> Create channel and join error:", err)
		os.Exit(-1)
	}

	// create chaincode lifecycle for each chaincode
	for _, cc := range chaincodes {
		info.ChaincodeID = cc.Name
		info.ChaincodePath = cc.Path
		info.ChaincodeVersion = cc_version

		if err := sdkInit.CreateCCLifecycle(&info, 1, false, sdk); err != nil {
			fmt.Printf(">> create chaincode lifecycle for %s error: %v\n", cc.Name, err)
			os.Exit(-1)
		}

		// invoke chaincode set status
		fmt.Printf(">> 通过链码外部服务设置链码状态，chaincode ID: %s ...\n", cc.Name)

		if err := info.InitService(info.ChaincodeID, info.ChannelID, info.Orgs[0], sdk); err != nil {
			fmt.Printf(">> InitService for %s successful\n", cc.Name)
			os.Exit(-1)
		}
		fmt.Printf(">> 设置链码状态完成，chaincode ID: %s\n", cc.Name)
	}

	App = sdkInit.Application{
		SdkEnvInfo: &info,
	}
	defer info.EvClient.Unregister(sdkInit.BlockListener(info.EvClient))
	defer info.EvClient.Unregister(sdkInit.ChainCodeEventListener(info.EvClient, info.ChaincodeID))

	// 定义链码路由和处理函数
	router := mux.NewRouter()

	chaincodeNames := []string{"nft", "image", "product"} // 替换为实际的链码名称列表
	
	//chaincodeNames := []string{"image", "product"} // 替换为实际的链码名称列表

	for _, chaincodeName := range chaincodeNames {
		registerChaincodeRoutes(router, chaincodeName)
	}

	// 使用CORS处理器包装路由器
	corsHandler := handlers.CORS(
		handlers.AllowedHeaders([]string{"Content-Type"}),
		handlers.AllowedMethods([]string{"GET", "POST"}),
		handlers.AllowedOrigins([]string{"*"}), // 这里使用通配符允许所有来源，可以根据需求进行设置
	)(router)
	log.Println("Listening on", "http://localhost:9000")
	err = http.ListenAndServe(":9000", corsHandler)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func registerChaincodeRoutes(router *mux.Router, chaincodeName string) {
	nftRouter := router.PathPrefix("/" + chaincodeName).Subrouter()
	nftRouter.HandleFunc("/CreateImageNFT", createImageNFTHandler).Methods("POST")
	nftRouter.HandleFunc("/GetImageNFTById", getImageNFTByIdHandler).Methods("POST")
	nftRouter.HandleFunc("/GetImageNFTByOR", getImageNFTByORHandler).Methods("POST")
	nftRouter.HandleFunc("/TransferImageNFT", transferImageNFTHandler).Methods("POST")
	nftRouter.HandleFunc("/BurnImageNFT", burnImageNFTHandler).Methods("POST")

	imageRouter := router.PathPrefix("/" + chaincodeName).Subrouter()
	imageRouter.HandleFunc("/CreateImage", createImageHandler).Methods("POST")
	imageRouter.HandleFunc("/GetImage", getImageHandler).Methods("POST")
	imageRouter.HandleFunc("/GetAllImages", getAllImagesHandler).Methods("POST")
	
	productRouter := router.PathPrefix("/" + chaincodeName).Subrouter()
	productRouter.HandleFunc("/CreateProduct", createProductHandler).Methods("POST")
	productRouter.HandleFunc("/GetProduct", getProductHandler).Methods("POST")
	productRouter.HandleFunc("/GetAllProducts", getAllProductsHandler).Methods("POST")

	/*
		// 注册获取通道信息的路由
		channelRouter := router.PathPrefix("/channel").Subrouter()
		channelRouter.HandleFunc("/GetChannelInfo", getChannelInfoHandler).Methods("POST")
		channelRouter.HandleFunc("/GetBlockNumberByTxID", getBlockNumberByTxIDHandler).Methods("POST")
	*/
}

// 创建product
func createProductHandler(w http.ResponseWriter, r *http.Request) {

	// 读取请求的参数
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 构造调用链码函数的参数
	var args []string
	args = append(args, "CreateProduct")
	args = append(args, r.Form.Get("id"))
	args = append(args, r.Form.Get("patientName"))
	args = append(args, r.Form.Get("localRoute"))
	args = append(args, r.Form.Get("modalCode"))
	args = append(args, r.Form.Get("checkTime"))
	args = append(args, r.Form.Get("status"))
	args = append(args, r.Form.Get("attachmentHash"))
	args = append(args, r.Form.Get("totalHash"))

	// 调用链码函数
	response, err := App.CreateProduct("product", args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 将字节数组转换为字符串
	txID := string(response)

	// 返回JSON格式的响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"txID": txID})
}

//查询product
func getProductHandler(w http.ResponseWriter, r *http.Request) {

	// 读取请求的参数
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 构造链码调用参数
	args := []string{
		"GetProduct",
		r.Form.Get("id"),
	}

	// 调用链码
	response, err := App.GetProduct("product", args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回链码函数调用结果
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}

// 查询所有的product
func getAllProductsHandler(w http.ResponseWriter, r *http.Request) {
	// 调用链码
	response, err := App.GetAllProducts("product")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回链码函数调用结果
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}

/*
//获取通道信息
func getChannelInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 构造返回的通道信息
	channelInfo := map[string]interface{}{
		"ChannelName":   info.ChannelID,
		"BlockCount":    sdkInit.GetBlockCount(App.SdkEnvInfo.ChClient),
		"TransactionCount": sdkInit.GetTransactionCount(App.SdkEnvInfo.ChClient),
	}

	// 将通道信息转换为JSON格式
	response, err := json.Marshal(channelInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回JSON格式的通道信息
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

//根据交易ID查询区块序号
func getBlockNumberByTxIDHandler(w http.ResponseWriter, r *http.Request) {
	// 读取请求的参数
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 获取交易ID
	txID := r.Form.Get("txID")

	// 查询区块序号
	blockNumber, err := sdkInit.GetBlockNumberByTxID(App.SdkEnvInfo.ChClient, txID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 构造返回的区块序号
	response := map[string]interface{}{
		"TxID":        txID,
		"BlockNumber": blockNumber,
	}

	// 将区块序号转换为JSON格式
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回JSON格式的区块序号
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
*/

// 创建NFT
func createImageNFTHandler(w http.ResponseWriter, r *http.Request) {
	// 读取请求的参数
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 构造调用链码函数的参数
	args := make([]string, 0)
	args = append(args, "CreateImageNFT")
	args = append(args, r.Form.Get("id"))
	args = append(args, r.Form.Get("owner"))
	args = append(args, r.Form.Get("data"))
	args = append(args, r.Form.Get("operationRecord"))
	args = append(args, r.Form.Get("digitalWatermark"))

	// 调用链码函数
	response, err := App.CreateImageNFT("nft", args) // 将链码名称作为参数传递，这里使用"nft"作为示例
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 将字节数组转换为字符串
	txID := string(response)

	// 返回JSON格式的响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"txID": txID})
}

//通过id查询NFT
func getImageNFTByIdHandler(w http.ResponseWriter, r *http.Request) {

	// 读取请求的参数
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 构造链码调用参数
	args := []string{
		"GetImageNFTById",
		r.Form.Get("id"),
	}

	// 调用链码
	response, err := App.GetImageNFTById("nft", args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回链码函数调用结果
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}

//通过操作记录查询NFT
func getImageNFTByORHandler(w http.ResponseWriter, r *http.Request) {

	// 读取请求的参数
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 构造链码调用参数
	args := []string{
		"GetImageNFTByOR",
		r.Form.Get("id"),
		r.Form.Get("operationRecord"),
	}

	// 调用链码
	response, err := App.GetImageNFTByOR("nft", args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回链码函数调用结果
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}

//转移NFT
func transferImageNFTHandler(w http.ResponseWriter, r *http.Request) {

	// 读取请求的参数
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 构造链码调用参数
	args := []string{
		"TransferImageNFT",
		r.Form.Get("id"),
		r.Form.Get("from"),
		r.Form.Get("to"),
	}

	// 调用链码
	response, err := App.TransferImageNFT("nft", args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 将字节数组转换为字符串
	txID := string(response)

	// 返回JSON格式的响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"txID": txID})
}

//销毁NFT
func burnImageNFTHandler(w http.ResponseWriter, r *http.Request) {

	// 读取请求的参数
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 构造链码调用参数
	args := []string{
		"BurnImageNFT",
		r.Form.Get("id"),
	}

	// 调用链码
	response, err := App.BurnImageNFT("nft", args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 将字节数组转换为字符串
	txID := string(response)

	// 返回JSON格式的响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"txID": txID})
}

// 创建image
func createImageHandler(w http.ResponseWriter, r *http.Request) {

	// 读取请求的参数
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 构造调用链码函数的参数
	var args []string
	args = append(args, "CreateImage")
	args = append(args, r.Form.Get("id"))
	args = append(args, r.Form.Get("patientName"))
	args = append(args, r.Form.Get("localRoute"))
	args = append(args, r.Form.Get("modalCode"))
	args = append(args, r.Form.Get("checkTime"))
	args = append(args, r.Form.Get("status"))
	args = append(args, r.Form.Get("attachmentHash"))
	args = append(args, r.Form.Get("totalHash"))

	// 调用链码函数
	response, err := App.CreateImage("image", args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 将字节数组转换为字符串
	txID := string(response)

	// 返回JSON格式的响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"txID": txID})
}

//查询image
func getImageHandler(w http.ResponseWriter, r *http.Request) {

	// 读取请求的参数
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 构造链码调用参数
	args := []string{
		"GetImage",
		r.Form.Get("id"),
	}

	// 调用链码
	response, err := App.GetImage("image", args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回链码函数调用结果
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}

// 查询所有的image
func getAllImagesHandler(w http.ResponseWriter, r *http.Request) {
	// 调用链码
	response, err := App.GetAllImages("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回链码函数调用结果
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}
