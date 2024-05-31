# testflow

Convert the `json` exported by `charles` into `qor5` test case references

```shell
go run ./*.go --sample-file ./integration_test/sample/New.json --patch-file ./integration_test/patch/New.origin --backup-dir ./integration_test/_backup --output-file ./integration_test/flow_new_test.go 

sh ./gen.sh ./integration_test/sample ./integration_test/patch ./integration_test/_backup ./integration_test
```
