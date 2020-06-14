<template>
    <div class="keys">
        <div class="path">
            <span @click="openKeyForPath({
                    key:'',
                    name:''
                })">
                <Icon type="ios-home-outline" /></span>
            <span v-show="showPathInput == false">
                <span v-for="(item,key) in paths" :key="key" @click="openKeyForPath(item)"><span
                        v-if="key > 0 && item.name != '/' ">/</span> {{ item.name }} </span>
            </span>

            <!-- 路径文本框 -->
            <Input v-show="showPathInput == true" @on-enter="enterPath" v-model="currentPath"
                :placeholder="$t('key.currentPathTips')" style="width: 300px;margin-top:-9px;font-size:28px;" />
            <span @click="showPathInput = !showPathInput" style="font-size:15px;margin-left:5px;">
                <Button type="primary" v-show="showPathInput == false" size="small">{{$t('public.edit')}}</Button>
                <Button type="primary" v-show="showPathInput == true" size="small">{{$t('public.hide')}}</Button>
            </span>

            <!-- 语言 -->
            <RadioGroup v-model="lang" @on-change="changeLang" type="button" style="float:right;margin-right:6px;">
                <Radio label="en">EN</Radio>
                <Radio label="zh">ZH</Radio>
            </RadioGroup>
            <!-- 展示方式切换 -->
            <RadioGroup v-model="listType" @on-change="changeListType" type="button"
                style="float:right;margin-right:6px;">
                <Radio label="list">
                    <Icon type="ios-list-box-outline" />
                </Radio>
                <Radio label="json"><img src="../assets/imgs/json.png" alt="json"
                        style="width:16px;height:auto;margin-top:7px"></Radio>
            </RadioGroup>

        </div>

        <div class="search">
            <Form inline>
                <FormItem>
                    <Select v-model="etcdName" style="text-align: left;width:300px" @on-change="changeEtcdName">
                        <Option v-for="item in etcdServers" :label="item.Title" :value="item.Name" :key="item.Name">
                            <b>{{ item.Title }}</b>
                            <span style="float:right;color:#ccc">{{ item.Name }}</span>
                        </Option>
                    </Select>
                </FormItem>
                <FormItem class="search-in">
                    <Input v-model="searchVal" type="text" @on-keyup="onSearchLocal">
                    <Button slot="prepend" type="primary">{{$t('public.screen')}}</Button>
                    <Button slot="append" type="primary" icon="ios-search" @click="onSearchLocal"></Button>
                    </Input>
                </FormItem>
            </Form>

        </div>

        <div>
            <Row>
                <Col span="6">
                <div class='tree'>
                    <Tree @on-toggle-expand='onToggleExpandTree' @on-select-change='onSelectChangeTree' :data="treeData">
                    </Tree>
                </div>
                
                </Col>


                <Col span="18">
                <div class="lists">
                    <!-- 表格形式展示 -->
                    <div class="table-list-main" v-if="listType == 'list'">
                        <Table border :columns="columnsKey" :data="keysList1" @on-selection-change="selectionChangeKeys"
                            :loading="keyLoading"></Table>
                        <div style="margin-top:10px; text-align: right;" v-if="pageShow == true">
                            <Page @on-change="changeListPage" @on-page-size-change="pageSizeChange" show-sizer
                                :current="page" :page-size="pageSize" :total="listTotal" />
                        </div>

                    </div>

                    <!-- json格式展示 -->
                    <div class="key-json-main" v-show="listType == 'json' || listType == 'toml' || listType == 'yaml'">
                        <codemirror v-model="mainConfig" :options="cmOptionShow" style="border: 1px solid #dcdee2;"
                            ref="listEditor"></codemirror>
                    </div>

                    <div style="height:200px;" v-show="!(listType == 'json' || listType == 'toml' || listType == 'yaml')">
                    </div>

                    <div style="clear:both"></div>

                </div>
                </Col>
            </Row>
        </div>



        <!-- 查看弹框 -->
        <Drawer :width="60" v-model="showKeyInfoModel" :title="$t('key.editKey')">
            <Form :model="showKeyInfo" :label-width="80">
                <FormItem label="Key" prop="key">
                    <Input v-model="showKeyInfo.full_dir" disabled placeholder="key"></Input>
                </FormItem>
                <FormItem label="Version" prop="version">
                    <Input v-model="showKeyInfo.version" disabled placeholder="Version"></Input>
                </FormItem>
                <FormItem label="Value" prop="value">
                    <codemirror v-model="showKeyInfo.value" :options="cmOption"
                        style="line-height:20px;border: 1px solid #dcdee2;" ref="showEditor"></codemirror>
                </FormItem>
                <FormItem>
                    <Button @click="saveKey" type="primary">{{$t('public.save')}}</Button>
                    <Button @click="showKeyInfoModel = false" style="margin-left: 8px">{{$t('public.close')}}</Button>
                </FormItem>

            </Form>
        </Drawer>
    </div>
</template>

<script>
    require("codemirror/mode/javascript/javascript");
    require("codemirror/mode/toml/toml");
    require("codemirror/mode/yaml/yaml");
    require("codemirror/mode/xml/xml");

    // require('codemirror/addon/hint/show-hint.css');
    // require('codemirror/addon/hint/show-hint');
    // require('codemirror/addon/hint/javascript-hint');
    // require('codemirror/addon/hint/anyword-hint');


    export default {
        data() {
            return {
                // json 形式展示
                mainConfig: '',

                treeData: [],
                listTotal: 0, // 总条数
                pageSize: 10, // 默认10条
                page: 1,
                pageShow: true, // 是否显示分页

                keyLoading: false, // 是否数据加载中

                etcdServers: [],

                lang: 'en',
                etcdName: '', // etcd服务名
                keyPrefix: '', // key 前缀

                searchVal: '', // 搜索内容

                showPathInput: false, // 是否显示路径文本框
                listType: "list", // 显示方式
                paths: [{
                    key: "/",
                    name: ""
                }], // 路径
                currentPath: "", // 当前key路径
                showKeyInfoModel: false, // 配置详情是否显示
                showKeyInfo: {}, // 要显示的配置值

                selectionKeys: [], // 表格选中列表

                keysList: [],
                keysList1: [], // 分页使用

                columnsKey: [
                    // 表格形式展示表头
                    {
                        type: "selection",
                        width: 60,
                        align: "center"
                    },
                    {
                        title: "NAME",
                        key: "title"
                    },
                    {
                        title: "KEY",
                        key: "key"
                    },
                    {
                        title: "Version",
                        key: "version"
                    },
                    {
                        title: " ",
                        align: "center",
                        render: (h, params) => {
                            return h("div", [
                                h(
                                    "Button", {
                                        props: {
                                            type: "primary",
                                            size: "small"
                                        },
                                        style: {
                                            marginRight: "5px"
                                        },
                                        on: {
                                            click: () => {
                                                this.openKey(params.row);
                                            }
                                        }
                                    },
                                    params.row.is_dir == true ? this.$t('key.open') : this.$t(
                                        'key.show')
                                ),

                            ]);
                        }
                    }
                ],
                // 代码编辑器配置
                cmOption: {
                    tabSize: 4,
                    smartIndent: true,
                    styleActiveLine: true,
                    lineNumbers: true,
                    line: true,
                    mode: 'text/javascript',
                    lineWrapping: true,
                    theme: 'default',
                    // lint: true,
                    // gutters: ['CodeMirror-lint-markers'],
                },
                // 显示指定格式
                cmOptionShow: {
                    readOnly: 'nocursor',
                    tabSize: 4,
                    smartIndent: true,
                    styleActiveLine: true,
                    lineNumbers: true,
                    line: true,
                    mode: 'text/javascript',
                    lineWrapping: true,
                    theme: 'default'
                }

            };
        },
        mounted() {
            this.getEtcdServers();

            this.lang = localStorage.getItem('lang') || 'en';
            this.listType = localStorage.getItem("list_type") || 'list';
            this.currentPath = this.keyPrefix;
            this.etcdName = localStorage.getItem("etcdName") || '';

            // 编辑器高度
            this.$refs.addEditor.codemirror.setSize('auto', '60vh');
            this.$refs.showEditor.codemirror.setSize('auto', '60vh');
            this.$refs.listEditor.codemirror.setSize('auto', '75vh');

        },
        methods: {
            // 路径文本框回车
            enterPath() {
                this.openKeyForPath({
                    key: this.currentPath,
                    name: ""
                });
            },
            // 选中key
            checkKey(item) {
                item.check = !item.check;
            },
            // 打开key
            openKey(item) {
                console.log(item);
                if (item.is_dir == false) {
                    this.showKeyInfoModel = true;
                    // 查询key的值
                    this.$http.get(`/v1/key?key=${item.key}`, {
                        headers: {
                            "EtcdServerName": this.etcdName,
                        }
                    }).then(response => {
                        if (response.status == 200) {
                            this.showKeyInfo = response.data;
                            console.log(this.showKeyInfo)
                        }
                    }).catch(error => {
                        if (error.response) {
                            this.$Message.error(error.response.data.msg);
                        }
                    })
                } else {
                    this.currentPath = item.full_dir;
                    this.getKeyList();
                }
            },
            // 顶部路径打开目录
            openKeyForPath(item) {
                console.log(item)
                this.currentPath = item.key || this.keyPrefix;
                if (this.currentPath == "/" && this.currentPath != this.keyPrefix) {
                    this.currentPath = this.keyPrefix;
                }
                console.log(this.currentPath)

                this.getKeyList();
            },

            saveKey() {
                let putData = this.showKeyInfo;
                putData.is_dir = false;
                this.$http
                    .put(`/v1/key`, putData, {
                        headers: {
                            "EtcdServerName": this.etcdName,
                        }
                    })
                    .then(response => {
                        console.log(response);
                        if (response.status == 200) {
                            this.$Message.success(this.$t('key.saveSuccessfully'));
                            this.getKeyList();
                            this.showKeyInfoModel = false;
                        }
                    }).catch(error => {
                        if (error.response) {
                            this.$Message.error(error.response.data.msg);
                        }
                    });
            },

            // 表格选中行
            selectionChangeKeys(selections) {
                this.selectionKeys = selections;
                console.log(selections);
            },

            // 获取key列表
            getKeyList() {
                if (this.listType == 'json' || this.listType == 'toml' || this.listType == 'yaml') {
                    this.getKeyShowConfig();
                }

                this.keyLoading = true;
                this.$Loading.start();
                let k = this.currentPath;
                if (k == "" || k == "/") {
                    k = this.keyPrefix;
                }
                console.log(this.keyPrefix);
                console.log(k);
                this.baseList = [];
                this.keysList = [];
                this.keysList1 = [];
                this.$http.get(`/v1/lsdir?key=${k}`, {
                    headers: {
                        "EtcdServerName": this.etcdName,
                    }
                }).then(response => {
                    console.log(response);
                    if (response.status == 200) {
                        // this.baseList = response.data || [];
                        // this.keysList = response.data || [];
                        // this.listTotal = this.keysList.length;
                        // this.changeListPage(1);
                        // this.page = 1;
                        this.$Loading.finish();
                        this.makeTree(response.data)
                    } else {
                        this.$Loading.error();
                    }
                    this.keyLoading = false;
                }).catch(error => {
                    if (error.response) {
                        this.$Message.error(error.response.data.msg);
                    }
                    this.keyLoading = false;
                    this.$Loading.error();
                });
            },
            onSelectChangeTree(selected, current) {
                console.log(current)
                if (!current.is_dir) {
                    this.openKey(current)
                }
            },
            onToggleExpandTree(current) {
                console.log(current)
                if (current.expand) {
                    // 展开节点
                    let showInList = []
                    for (let child of current.children) {
                        if (!child.is_dir) {
                            showInList.push(child)
                        }
                    }
                    this.baseList = showInList
                    this.listTotal = showInList.length;
                    this.page = 1;
                    this.keysList1 = this.baseList.slice((this.page - 1) * this.pageSize, this.page * this.pageSize);
                    this.changeListPage(1);
                    
                    this.$Loading.finish();
                } else {
                    // 关闭节点
                }
            },
            makeTree(data) {
                console.log(111111111111111111111111)
                console.log(data)
                for (let node of data) {

                }
                this.treeData = data
            },

            // 搜索本地
            onSearchLocal(e) {
                if (e.keyCode == 13) {
                    return
                }
                console.log(this.searchVal)
                if (this.searchVal == '') {
                    this.pageShow = true;
                } else {
                    this.pageShow = false;
                }
                let list = [];
                if (typeof this.baseList == 'undefined') {
                    return
                }
                this.baseList.forEach(val => {
                    let key = val.key;
                    if (key.indexOf(this.searchVal) >= 0 || this.searchVal == '') {
                        list.push(val);
                    }
                });

                this.keysList1 = list;
                this.keysList = list;
                if (this.searchVal == '' && this.listType == 'list') {
                    this.changeListPage(1);
                    this.page = 1;
                }
            },

            // 展现方式
            changeListType() {
                localStorage.setItem("list_type", this.listType || 'list');
                if (this.listType == 'json' || this.listType == 'toml' || this.listType == 'yaml') {
                    this.getKeyShowConfig();
                    return
                }
                console.log(this.baseList)
                this.changeListPage(1);
                this.page = 1;

                this.onSearchLocal();
            },

            // 切换语言
            changeLang() {
                this.$i18n.locale = this.lang || 'en';
                localStorage.setItem("lang", this.lang || 'en');
            },

            // 切换页码
            changeListPage(page) {
                // pageSize
                this.keysList1.splice(0, this.keysList1.length);
                this.keysList1 = this.baseList.slice((page - 1) * this.pageSize, page * this.pageSize);
                console.log(page);
            },
            // 页面打小
            pageSizeChange(pageSize) {
                this.pageSize = pageSize;
                this.changeListPage(1);
                this.page = 1;
            },

            // 获取etcd server列表
            getEtcdServers() {
                this.$http.get(`/v1/server`).then(response => {
                    console.log(response);
                    if (response.status == 200) {
                        this.etcdServers = response.data || [];
                        if (this.etcdServers.length > 0) {
                            if (this.etcdName == '') {
                                this.etcdName = this.etcdServers[0].Name;
                                this.keyPrefix = this.etcdServers[0].KeyPrefix;
                                this.currentPath = this.keyPrefix;
                            } else {
                                this.etcdServers.forEach(val => {
                                    if (val.Name == this.etcdName) {
                                        this.keyPrefix = val.KeyPrefix;
                                        this.currentPath = this.keyPrefix;
                                    }
                                });
                            }
                            this.getKeyList();
                        }
                        localStorage.setItem("etcdName", this.etcdName)
                        console.log(this.etcdServers)
                    }
                }).catch(error => {
                    if (error.response) {
                        this.$Message.error(error.response.data.msg);
                    }
                });
            },

            // 切换服务端
            changeEtcdName(val) {
                console.log(val);
                this.etcdName = val;
                this.etcdServers.forEach(v => {
                    if (v.Name == val) {
                        this.keyPrefix = v.KeyPrefix;
                    }
                });
                this.currentPath = this.keyPrefix;
                this.getKeyList();
                localStorage.setItem("etcdName", this.etcdName)
            },

            // 获取当前key的json展现方式
            getKeyShowConfig() {
                this.$Loading.start();
                this.$http.get(`/v1/key/format?format=${this.listType}&key=${this.currentPath}`, {
                    headers: {
                        "EtcdServerName": this.etcdName,
                    }
                }).then(response => {
                    if (response.status == 200) {
                        this.mainConfig = response.data;
                        console.log(response)
                        this.$Loading.finish();
                    } else {
                        this.$Loading.error();
                    }
                }).catch(error => {
                    if (error.response) {
                        this.$Message.error(error.response.data.msg);
                    }
                    this.$Loading.error();
                });
            },


        },
        watch: {
            currentPath(newCurrentPath) {
                let paths = newCurrentPath.split("/");
                if (paths.length == 0) {
                    paths = [];
                    paths.push(this.keyPrefix);
                }
                console.log(paths)
                let fullDir = '';
                this.paths = [];
                paths.forEach(val => {
                    console.log(this.keyPrefix.indexOf("/"))
                    if (this.keyPrefix.indexOf("/") == 0) {
                        if (fullDir != '/') {
                            fullDir += '/';
                        }
                    } else if (fullDir != '') {
                        fullDir += '/';
                    }
                    fullDir += val;
                    this.paths.push({
                        key: fullDir, // .trim('/')
                        name: val
                    });
                    console.log(this.paths)
                });

            }
        }
    };
</script>

<style scoped>
    .keys {
        width: 100%;
        height: 100vh;
        overflow-y: scroll;
        overflow-x: hidden;
    }

    .keys .path {
        width: 100%;
        font-size: 24px;
        border-bottom: 1px solid #cecece;
    }

    .keys .path span {
        margin: 5px 0px 13px 0px;
        color: #333333;
        cursor: pointer;
    }

    .tree {
        overflow-x: auto;
    }

    .keys .lists {
        position: relative;
        width: 100%;
        height: 100vh;
        /* background-color: aqua; */
    }

    .keys .lists .key-list-main {
        position: absolute;
        left: 0;
        top: 0;
    }

    .keys .lists .key-list {
        position: relative;
        width: 110px;
        height: 110px;
        padding: 15px;
        margin: 10px;
        float: left;
        text-align: center;
        background-color: #eeeeee;
    }

    .keys .lists .key-list .title {
        font-size: 18px;
        width: 60px;
        overflow: hidden;
        /* text-overflow:ellipsis; */
        white-space: nowrap;
    }

    .keys .lists .key-list .checkbox {
        position: absolute;
        left: 10px;
        top: 10px;
    }

    .keys .lists .key-icon {
        width: 60px;
        height: 60px;
    }

    .keys .lists .table-list-main {
        margin-top: 10px;
    }

    .ivu-poptip-body-message {
        display: inline-block !important;
    }

    .search {
        margin-top: 10px;
        width: 100%;
        height: 36px;
        text-align: center;
    }

    .search .search-in {
        width: 50%;
        min-width: 300px;
        margin: 0 auto;
    }
</style>