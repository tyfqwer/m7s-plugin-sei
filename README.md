插入SEI接口地址
http://192.168.1.11:8080/sei/api/insertsei?streamPath=live/test

#
安装 go get github.com/tyfqwer/m7s-plugin-sei

# 打包
goreleaser release --skip=publish --skip-validate

# 安装
go install github.com/goreleaser/goreleaser@latest
