package database

import (
	"context"
	"time"
	"wakeUpCall/utils"

	"cloud.google.com/go/firestore"
)

var PeerCollectionName = "peers"

type Peer struct{
	PeerID string `json:"peer_id"`
	//interface{}に注意。
	//TimeScheduleは，場合によって文字列で扱いたいときと，time.Timeで扱いたい時があるから。
	TimeSchedule interface{} `json:"time_schedule"`
	Comment string `json:"comment"`
	IsWaiting bool `json:"isWaiting"`
}

func NewPeer(peerID string,timeSchedule interface{},comment string, isWaiting bool) *Peer{
	return &Peer{
		peerID,
		timeSchedule,
		comment,
		isWaiting,
	}
}

//予約する（ドキュメントを追加する）
func (c *Peer)Update()error{
	ctx:=context.Background()

	//初期化処理
	client,err:=FirebaseInit(ctx)
	if err!=nil{
		return err
	}

	//c.TimeSchedule = utils.GMTToJST(c.TimeSchedule.(time.Time))

	//データベースに追加
	_,err = client.Collection(PeerCollectionName).Doc(c.PeerID).Set(ctx,map[string]interface{}{
		"time_schedule":c.TimeSchedule,
		"comment":c.Comment,
		"isWaiting":c.IsWaiting,
	})
	if err!=nil{
		return err
	}

	defer client.Close()

	return nil
}

//ステータスチェック
func CheckStatus(peerID string) (interface{}, error){
	ctx:=context.Background()

	//初期化処理
	client,err:=FirebaseInit(ctx)
	if err!=nil{
		return nil,err
	}

	//ステータス取得
	dsnap,err:=client.Collection(PeerCollectionName).Doc(peerID).Get(ctx)
	if err != nil{
		//falseは仮置き
		return nil,err
	}
	m:=dsnap.Data()

	defer client.Close()

	return m["isWaiting"],nil
}

//isWaitingがtrue(つまりwaiting)である，すべてのPeerを取得
func GetWaitingPeers()([]*Peer,error){
	ctx:=context.Background()

	//初期化処理
	client,err:=FirebaseInit(ctx)
	if err!=nil{
		return nil,err
	}



	//nowGMTtime:=utils.JSTtoGMT(time.Now())
	//今から20分後の時刻
	//inTwentyMinutes:=time.Now().Add(20*time.Minute)

	//herokuでは，デフォルトのタイムゾーンがPST
	nowJSTtime:=utils.PSTtoJST(time.Now())
	//JSTが，現在時刻(JST)から9時間遅れていることを考慮
	databaseTime:=nowJSTtime.Add(9*time.Hour)
	inTwentyMinutes:=databaseTime.Add(20*time.Minute)
	//データ読み込み
	//必ず，CloudFirestoreにて複合インデックスを作成してから，実行すること。
	//作成していない場合，次のエラーメッセージが表示されると思われる。
	//その場合，'You can create it here'の後に記述されているurlにアクセスし，指示に従って複合インデックスを作成すること。
	//2020/09/08 03:46:27 webserver.go:56: action=showTimeline, err= rpc error: code = FailedPrecondition desc = The query requires an index. You can create it here: https://console.firebase.google.com/v1/r/project/{...}
	//なお，設定時刻が過ぎたPeerは自動的に着信される(はず)ため，time.Now()<=は本当はいらないかも
	allData:=client.Collection(PeerCollectionName).Where("time_schedule","<=",inTwentyMinutes).Where("time_schedule",">=",databaseTime).Where("isWaiting","==",true).OrderBy("time_schedule",firestore.Asc).Documents(ctx)
	//allData:=client.Collection(PeerCollectionName).Where("isWaiting","==",true).Documents(ctx)
	//ドキュメント取得
	docs,err:=allData.GetAll()
	if err!=nil{
		return nil,err
	}

	//配列の初期化
	peers:=make([]*Peer,0)
	for _,doc:=range docs{
		//構造体の初期化
		p:=new(Peer)

		mapPeer:=doc.Data()
		//ドキュメント名を取得して，PeerIDにセット
		p.PeerID=doc.Ref.ID

		p.TimeSchedule=mapPeer["time_schedule"].(time.Time)
		//タイムゾーンをUTC→JSTに直す。
		p.TimeSchedule=utils.UTCtoJST(p.TimeSchedule.(time.Time))
		//タイムゾーンをJST→MGTに直す
		p.TimeSchedule=utils.JSTtoGMT(p.TimeSchedule.(time.Time))

		//さらに，time.Time→stringに直す。
		p.TimeSchedule=utils.TimeToString(p.TimeSchedule.(time.Time))
		//さらにさらに，'hh:mm'だけ切り出し，TimeScheduleにセット
		p.TimeSchedule=utils.ToHHMM(p.TimeSchedule.(string))
		//Comment,IsWaitingにセット
		p.Comment=mapPeer["comment"].(string)
		p.IsWaiting=mapPeer["isWaiting"].(bool)

		//配列に構造体をセット
		peers=append(peers,p)
	}

	defer client.Close()

	return peers,nil
}