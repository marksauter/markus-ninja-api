package utils

import (
	"os"

	"github.com/joho/godotenv"
)

// var homoglyphs map[rune]struct{}
//
// func init() {
//   homoglyphs = make(map[rune]struct{})
//   for _, r := range `ᅟᅠ         　ㅤǃ！״″＂＄％＆＇﹝（﹞）⁎＊＋‚，‐𐆑－٠۔܁܂․‧。．｡⁄∕╱⫻⫽／ﾉΟοОоՕ𐒆ＯｏΟοОоՕ𐒆Ｏｏا１２３４５６𐒇７Ց８９։܃܄∶꞉：;；‹＜𐆐＝›＞？＠［＼］＾＿｀ÀÁÂÃÄÅàáâãäåɑΑαаᎪＡａßʙΒβВЬᏴᛒＢｂϲϹСсᏟⅭⅽ𐒨ＣｃĎďĐđԁժᎠⅮⅾＤｄÈÉÊËéêëĒēĔĕĖėĘĚěΕЕеᎬＥｅϜＦｆɡɢԌնᏀＧｇʜΗНһᎻＨｈɩΙІіاᎥᛁⅠⅰ𐒃ＩｉϳЈјյᎫＪｊΚκКᏦᛕKＫｋʟιاᏞⅬⅼＬｌΜϺМᎷᛖⅯⅿＭｍɴΝＮｎΟοОоՕ𐒆ＯｏΟοОоՕ𐒆ＯｏΡρРрᏢＰｐႭႳＱｑʀԻᏒᚱＲｒЅѕՏႽᏚ𐒖ＳｓΤτТᎢＴｔμυԱՍ⋃ＵｕνѴѵᏙⅤⅴＶｖѡᎳＷｗΧχХхⅩⅹＸｘʏΥγуҮＹｙΖᏃＺｚ｛ǀا｜｝⁓～ӧӒӦ` {
//     homoglyphs[r] = struct{}{}
//   }
// }

func LoadEnv() error {
	// Load env vars from .env
	return godotenv.Load()
}

func GetOptionalEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetRequiredEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic("Env variable " + key + " required.")
}

// func IsHomoglyph(r rune) bool {
//   _, ok := homoglyphs[r]
//   return ok
// }
