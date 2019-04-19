package jwt

import (
  "f5wallet/server/config"
	"errors"
	"fmt"
	// "github.com/valyala/fasthttp"
	"github.com/qiangxue/fasthttp-routing"
	  "encoding/json"
)

type ApiResponse struct  {
	  Rescode int  `json:"rescode"`
		Resdecr string  `json:"resdecr"`
		Resdata interface{}  `json:"resdata"`
}

func (res *ApiResponse) ToJson() string {
	  b, err := json.Marshal(res)
    if err != nil {
        //fmt.Println("Json parse error: : ",err)
        return ""
    }
		//fmt.Println("Json parse ok: : ",string(b))
    return string(b)
}

func CheckTokenMiddleware(c *routing.Context) error {

	cfg := config.GetConfig()
	if cfg.Jwt.Enable {
		fasthttpJwtCookie := c.Request.Header.Cookie("fasthttp_jwt")

		if len(fasthttpJwtCookie) == 0 {
			return errors.New("login required")
		}

		JWTSignKey :=  cfg.Jwt.Signkey
		token, _, err := JWTValidate(JWTSignKey, string(fasthttpJwtCookie))

		if !token.Valid {
			return errors.New("your session is expired, login again please")
		}
		return err
	}
	return nil
}

type MiddlewareType func(c *routing.Context) error

var MiddlewareList = []MiddlewareType{
	CheckTokenMiddleware,
}

// BasicAuth is the basic auth handler
func JWTMiddleware(next routing.Handler) routing.Handler {
	return routing.Handler(func(c *routing.Context) error {
		for _, middleware_item := range MiddlewareList {
			if err := middleware_item(c); err != nil {
				res := &ApiResponse{
						Rescode: 99,
						Resdecr: "Please login",
						Resdata: err.Error(),
				}
				fmt.Fprintf(c,res.ToJson())
				return nil
			}
		}
		return next(c)
	})
}
