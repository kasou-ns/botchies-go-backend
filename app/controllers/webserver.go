package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"wakeUpCall/app/database"
	"wakeUpCall/config"
	"wakeUpCall/utils"
)

type JSONError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

//API(バックエンド処理)にエラー発生
func APIError(w http.ResponseWriter, errMessage string, code int) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "https://mezamashi-jinji.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	//レスポンスにエラーコードを含める
	w.WriteHeader(code)
	jsonError, err := json.Marshal(JSONError{Error: errMessage, Code: code})
	if err != nil {
		log.Println(err)
	}
	w.Write(jsonError)
}

var apiValidPath = regexp.MustCompile("/(reserve|show_timeline|check_status|incoming)")

//HandleFuncの引数によく合います
//おいしいよ
func apiMakeHandler(fn func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := apiValidPath.FindStringSubmatch(r.URL.Path)
		if len(m) == 0 {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			//w.Header().Set("Access-Control-Allow-Origin", "https://mezamashi-jinji.vercel.app")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")

			if r.Method=="OPTIONS"{
				w.WriteHeader(http.StatusOK)
				return
			}
			http.NotFound(w, r)
			return
		}
		fn(w, r)
	}
}

//タイムラインを表示する
func showTimelineHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "https://mezamashi-jinji.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")

	if r.Method=="OPTIONS"{
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		APIError(w, "method is not POST", http.StatusBadRequest)
		return
	}

	//ステータスがtrueであるデータを全取得
	Peers, err := database.GetWaitingPeers()
	if err != nil {
		log.Println("action=showTimeline, err=", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//レスポンスのbodyにデータを加える
	js, err := json.Marshal(Peers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	w.WriteHeader(http.StatusOK)
}

//予約する(アラームをセットする)
func reserveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "https://mezamashi-jinji.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")

	if r.Method=="OPTIONS"{
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		APIError(w, "method is not POST", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "application/json"{
		APIError(w,"not application/json",http.StatusBadRequest)
		return
	}

	//To allocate slice for request body
	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Read body data to parse json
	body := make([]byte, length)
	length, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//parse json
	var jsonBody map[string]interface{}
	err = json.Unmarshal(body[:length], &jsonBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//パラメータ解析
	//エンコードタイプを，フロント側と確認すること
	//以下では，x-www-form-urlencodedと仮定
	/*
	err := r.ParseForm()
	if err != nil {
		APIError(w, "Not x-www-form-urlencoded", http.StatusBadRequest)
		return
	}
	 */

	//log.Println(r)

	//peerIDの取得
	peerID := jsonBody["peer_id"]
	if peerID == "" {
		APIError(w, "No peer_id param", http.StatusBadRequest)
		return
	}

	//予約時刻の取得
	strTimeSchedule := jsonBody["time_schedule"]
	timeSchedule,err := utils.TimeParser(strTimeSchedule.(string))
	if err != nil{
		APIError(w,"cannot parse time_schedule",http.StatusBadRequest)
		return
	}

	timeSchedule = utils.GMTToJST(timeSchedule)

	//コメントの取得
	//commentは空文字列もあり得る，エラーハンドリングなし
	comment := jsonBody["comment"]

	Peer := database.NewPeer(peerID.(string), timeSchedule, comment.(string), true)

	//データベースに追加(peerID,timeSchedule,comment)
	err = Peer.Update()
	if err != nil {
		log.Println("action=reserve,err=", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//予約完了
	w.WriteHeader(http.StatusOK)
}

//ステータスの取得
func checkStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "https://mezamashi-jinji.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")

	if r.Method=="OPTIONS"{
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		APIError(w, "method is not GET", http.StatusBadRequest)
		return
	}

	//peerIDの取得
	peerID := r.URL.Query().Get("peer_id")
	if peerID == "" {
		APIError(w, "No peer_id param", http.StatusBadRequest)
		return
	}

	//データベースからステータスを取得
	isWaitingInterface, err := database.CheckStatus(peerID)
	isWaiting := isWaitingInterface.(bool)
	if err != nil {
		log.Println("action=checkStatus, err=", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//isWaitingをjsonにしてbodyに含める
	js, err := json.Marshal(map[string]bool{
		"isWaiting": isWaiting,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	w.WriteHeader(http.StatusOK)
}

//着信を処理する
//データベースのステータスを，true→falseに書き換える
func incomingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "https://mezamashi-jinji.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")

	if r.Method=="OPTIONS"{
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		APIError(w, "method is not POST", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "application/json"{
		APIError(w,"not application/json",http.StatusBadRequest)
		return
	}

	//To allocate slice for request body
	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Read body data to parse json
	body := make([]byte, length)
	length, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//parse json
	var jsonBody map[string]interface{}
	err = json.Unmarshal(body[:length], &jsonBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	Peer := database.NewPeer(jsonBody["peer_id"].(string), "", "", false)

	//データベースに追加(peerID,timeSchedule,comment)
	err = Peer.Update()
	if err != nil {
		log.Println("action=reserve,err=", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//着信処理完了
	w.WriteHeader(http.StatusOK)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "https://mezamashi-jinji.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	if r.Method=="OPTIONS"{
		w.WriteHeader(http.StatusOK)
		return
	}

	io.WriteString(w, "Hello, world!")
}

func StartWebServer() error {
	port := os.Getenv("PORT") //heroku用
	http.HandleFunc("/",hello)
	http.HandleFunc("/show_timeline", apiMakeHandler(showTimelineHandler))
	http.HandleFunc("/reserve", apiMakeHandler(reserveHandler))
	http.HandleFunc("/check_status", apiMakeHandler(checkStatusHandler))
	http.HandleFunc("/incoming", apiMakeHandler(incomingHandler))
	http.ListenAndServe(":"+port, nil) //heroku用
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), nil)
}
