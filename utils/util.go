package utils

import (
	"log"
	"time"
)

func GMTToJST(timeSchedule time.Time)time.Time{
	//タイムゾーンを決める
	jst:=time.FixedZone("JST",9*60*60)
	//与えられた時刻のタイムゾーンについて，UTC→JSTにする
	timeSchedule = timeSchedule.In(jst)
	return timeSchedule
}

func UTCtoJST(timeSchedule time.Time)time.Time{
	//タイムゾーンを決める
	jst:=time.FixedZone("JST",9*60*60)
	//与えられた時刻のタイムゾーンについて，UTC→JSTにする
	timeSchedule = timeSchedule.In(jst)
	return timeSchedule
}

func JSTtoGMT(timeSchedule time.Time)time.Time{
	//タイムゾーンを決める
	//タイムゾーン，UTCとの時差を指定
	gmt:=time.FixedZone("GMT",0)
	//与えられた時刻のタイムゾーンについて，UTC→JSTにする
	timeSchedule = timeSchedule.In(gmt)
	return timeSchedule
}

func PSTtoJST(timeSchedule time.Time)time.Time{
	//タイムゾーンを決める
	jst:=time.FixedZone("JST",9*60*60)
	//与えられた時刻のタイムゾーンについて，UTC→JSTにする
	timeSchedule = timeSchedule.In(jst)
	return timeSchedule
}

func TimeToString(timeSchedule time.Time)string{
	return timeSchedule.Format("2006-01-02 15:04:05 (MST)")
}

func ToHHMM(timeSchedule string)string{
	return timeSchedule[11:16]
}

func TimeParser(strTimeSchedule string)(time.Time,error){
	tmp:=strTimeSchedule+" (GMT)"
	//パース対象が，どのように参照されているかのフォーマット定義
	layout:="2006-01-02 15:04:05 (MST)"
	timeSchedule,err:=time.Parse(layout,tmp)
	log.Println(timeSchedule)
	return timeSchedule,err
}