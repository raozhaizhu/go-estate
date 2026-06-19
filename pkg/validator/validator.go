package validator

import (
	"log"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// 全局翻译器暴露
var Trans ut.Translator

// InitTrans 初始化翻译器
func InitTrans() error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		zhT := zh.New()
		uni := ut.New(zhT, zhT)
		Trans, _ = uni.GetTranslator("zh")
		// 注册翻译器
		return zh_translations.RegisterDefaultTranslations(v, Trans)
	}
	log.Fatal("翻译器初始化失败")
	return nil
}
