给zimukux 0.2.0 做ocr服务，内部用腾讯云api

go install github.com/ccxp/ocr_for_zimukux@latest

ocr_zumukux  -l :80 -i secretId -k secretKey -r region
