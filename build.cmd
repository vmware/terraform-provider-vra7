cd S:\gosrc\src\github.com\cars\terraform-provider-vra7-1
go build
copy terraform-provider-vra7.exe /Y ..\test\example\simple\.terraform\plugins\windows_amd64
pushd S:\gosrc\src\github.com\cars\test\example\simple
del log
del vra-terraform.log
terraform init
terraform refresh
popd