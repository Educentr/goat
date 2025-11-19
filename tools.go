package goat

import "github.com/gin-gonic/gin"

func CheckURLParamStr(c *gin.Context, qname string) *string {
	id, ok := c.GetQuery(qname)
	if !ok {
		return nil
	}
	return &id
}
