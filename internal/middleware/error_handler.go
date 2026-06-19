package middleware

// func ErrorHandler() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		ctx.Next() // 放行, 执行后续逻辑

// 		if len(ctx.Errors) > 0 {
// 			ginErr := ctx.Errors.Last() // 获取最后发生的错误
// 			switch ginErr.Type {
// 			case gin.ErrorTypeBind: //是绑定错误
// 				response.FailWithBindError(ctx, ginErr.Err)
// 			default: // 是自定义错误或者内部错误
// 				response.Fail(ctx, ginErr.Err)
// 			}
// 			// 已处理错误
// 			ctx.Abort()
// 		}

// 	}
// }
