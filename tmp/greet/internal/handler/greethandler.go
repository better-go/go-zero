package handler

import (
	"net/http"

	"greet/internal/logic"
	"greet/internal/svc"
	"greet/internal/types"

	"github.com/tal-tech/go-zero/rest/httpx"
)

func greetHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Request
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}


		//
		// todo x: logic 是业务对象, 这部分用 logic 而不是 service, 好评. 避免概念混淆.
		//
		l := logic.NewGreetLogic(r.Context(), ctx)

		//
		// todo x: 业务方法调用
		//
		resp, err := l.Greet(req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			// todo x: rest json 返回值格式化
			httpx.OkJson(w, resp)
		}
	}
}
