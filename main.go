package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ocr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ocr/v20181119"
)

var listen, secretId, secretKey, region string

func init() {
	flag.StringVar(&listen, "l", "", "service listen address, e.g. ':11111'")
	flag.StringVar(&secretId, "i", "", "secretId, 使用腾讯云，密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取")
	flag.StringVar(&secretKey, "k", "", "secretKey")
	flag.StringVar(&region, "r", "", "region, 见 https://cloud.tencent.com/document/api/866/33518#.E5.9C.B0.E5.9F.9F.E5.88.97.E8.A1.A8")
	flag.Parse()
}

var ocrcli *ocr.Client

func OcrText(s string) (string, error) {

	data := s
	if p := strings.Index(s, ";base64,"); p >= 0 {
		data = s[p+len(";base64,"):]
	}

	request := ocr.NewEnglishOCRRequest()
	request.ImageBase64 = &data

	// 返回的resp是一个GeneralBasicOCRResponse的实例，与请求对象对应
	response, e := ocrcli.EnglishOCR(request)
	if e != nil {
		return "", e
	}

	var ret string
	if response.Response != nil {
		for _, r := range response.Response.TextDetections {
			if r.DetectedText != nil {
				ret += strings.ReplaceAll(*r.DetectedText, " ", "")
			}
		}
	}
	return ret, nil
}

type OcrRequest struct {
	ImageData string `json:"imgdata"`
}

type OcrResponse struct {
	Code   int    `json:"code"`
	Result string `json:"result"`
}

type OcrHandler struct {
}

func (h *OcrHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var e error
	defer func() {
		if e != nil {
			log.Printf("%v", e)
			http.Error(w, "404 page not found", http.StatusNotFound)
		}
	}()

	b, e := io.ReadAll(r.Body)
	if e != nil {
		return
	}

	req := OcrRequest{}
	e = json.Unmarshal(b, &req)
	if e != nil {
		return
	}

	ret, e := OcrText(req.ImageData)
	// log.Printf("%+v", req)
	// log.Printf("%v %v", ret, e)

	if e != nil {
		return
	}

	resp := OcrResponse{
		Code:   1,
		Result: ret,
	}
	if len(ret) == 0 {
		resp.Code = 0
	}

	b, e = json.Marshal(resp)
	if e != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.Write(b)
}

func main() {

	if listen == "" || secretId == "" || secretKey == "" || region == "" {
		flag.PrintDefaults()
		return
	}

	// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
	// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考，建议采用更安全的方式来使用密钥，请参见：https://cloud.tencent.com/document/product/1278/85305
	// 密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取
	credential := common.NewCredential(
		secretId,
		secretKey,
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ocr.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	ocrcli, _ = ocr.NewClient(credential, region, cpf)

	http.Handle("/ocr", new(OcrHandler))
	log.Fatal(http.ListenAndServe(listen, nil))
}
