每一步后都应该先自行记录前置种子文件，以及最终补充 交互和数据 的确认逻辑
- New: 新建
- Duplicate: 复制，再次复制
- VersionDialog: 打开版本列表，切换选择，切换tab，切换选择，关键词A，关键词B，选中当前显示并确认选择，选中非当前显示并确认选择
- PublishAndUnPulish: 发布，取消发布
- Schedule: 草稿态的 start < end < now，now < start < end，start < now < end，end < start < now ，其他态应代码调整


## DeleteVersion
1. 打开版本列表
2. 删除非当前选中也非当前显示
3. 选中其他项并删除，以测试删除当前选中是否会回选当前显示
4. 选中其他项但删除当前显示，测试其是否会切换显示但不更改选中
5. 删除当前显示并为当前选中，测试其是否会切换显示和选中为更老一个版本
6. 选中最老的版本并确认显示，然后再删除，以测试其在无更老版本切换的时候是否会切换和显示最新版本
7. 删除所有版本，测试其是否会直接返回至资源列表页

## Schedule
1. n < s < e ，成功
2. s < e < n ，报错
3. 两个记录即可，其他的情况复用前俩记录即可

```
# gen command
sh $GOPATH/src/github.com/qor5/admin/utils/testflow/gentool/gen.sh ./sample ./patch ./_backup .
```
