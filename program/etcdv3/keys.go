package etcdv3

import (
	"context"
	"errors"
	"path"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

func (c *Etcd3Client) LsDir(key string) (nodes []*Node, err error) {
	if key == "" {
		return nil, errors.New("key is empty")
	}
	dir := key
	if key != "/" {
		key = strings.TrimRight(key, "/")
		dir = key + "/"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	kv := clientv3.NewKV(c.Client)

	// 前缀查询,只读key
	resp, err := kv.Get(ctx, dir, clientv3.WithPrefix(),
		clientv3.WithKeysOnly(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	if err != nil {
		return nil, err
	} else {
		return c.generateTree(dir, resp.Kvs)
	}
}

func (c *Etcd3Client) generateTree(dir string, kvs []*mvccpb.KeyValue) (nodes []*Node, err error) {
	set := make(map[string]*Node)
	for _, kv := range kvs {
		name := strings.TrimPrefix(string(kv.Key), dir)

		nl := strings.Split(name, "/")
		nLength := len(nl)
		for i, n := range nl {
			var dirname string
			var key string

			dirname = dir + strings.Join(nl[0:i], "/")
			key = dir + strings.Join(nl[0:i+1], "/")
			parent := set[dirname]
			if parent == nil {
				if i == nLength-1 {
					parent = NewLeafNode(dirname, kv)
				} else {
					parent = NewNode(dirname, dirname)
				}

				set[dirname] = parent
				if i == 0 {
					nodes = append(nodes, parent)
				}
			}
			cur := set[key]
			if cur == nil {
				if i == nLength-1 {
					cur = NewLeafNode(key, kv)
				} else {
					cur = NewNode(key, n)
				}

				set[key] = cur
				parent.Children = append(parent.Children, cur)
			}

		}

	}

	return nodes, nil
}

// List 获取目录下列表
func (c *Etcd3Client) List(key string) (nodes []*Node, err error) {
	if key == "" {
		return make([]*Node, 0), errors.New("key is empty")
	}
	// 兼容key前缀设置为 /
	dir := key
	if key != "/" {
		key = strings.TrimRight(key, "/")
		dir = key + "/"
	}
	ctx := context.Background()
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()

	kv := clientv3.NewKV(c.Client)

	resp, err := kv.Get(ctx, dir, clientv3.WithPrefix())

	if err != nil {
		return nil, err
	} else {
		return c.list(dir, resp.Kvs)
	}
}

// Value 获取一个key的值
func (c *Etcd3Client) Value(key string) (val *Node, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.Client.Get(ctx, key)
	if err != nil {
		return
	}
	if resp.Kvs != nil && len(resp.Kvs) > 0 {
		val = &Node{
			Value:   string(resp.Kvs[0].Value),
			Version: resp.Kvs[0].Version,
		}
	} else {
		err = ErrorKeyNotFound
	}
	return
}

func (c *Etcd3Client) list(dir string, kvs []*mvccpb.KeyValue) ([]*Node, error) {
	nodes := []*Node{}
	for _, kv := range kvs {

		name := strings.TrimPrefix(string(kv.Key), dir)
		if strings.Contains(name, "/") {
			// secondary directory
			continue
		}
		nodes = append(nodes, NewLeafNode(dir, kv))
	}
	return nodes, nil
}

func (c *Etcd3Client) ensureKey(key string) (string, string) {
	key = strings.TrimRight(key, "/")
	if key == "" {
		return "/", ""
	}
	if strings.Contains(key, "/") == true {
		return key, path.Clean(key + "/../")
	} else {
		return key, ""
	}
}

// Put 添加一个key
func (c *Etcd3Client) Put(key string, value string, mustEmpty bool) error {
	// log.Println(key)
	key, _ = c.ensureKey(key)
	//  需要判断的条件
	cmp := make([]clientv3.Cmp, 0)


	if mustEmpty {
		cmp = append(
			cmp,
			clientv3.Compare(
				clientv3.Version(key),
				"=",
				0,
			),
		)
	} else {
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	txn := c.Client.Txn(ctx)
	// make sure the parentKey is a directory
	txn.If(
		cmp...,
	).Then(
		clientv3.OpPut(key, value),
	)

	txnResp, err := txn.Commit()
	if err != nil {
		return err
	}

	if !txnResp.Succeeded {
		return ErrorPutKey
	}
	return nil
}

// Delete 删除key
func (c *Etcd3Client) Delete(key string) error {
	key = strings.TrimRight(key, "/")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	txn := c.Client.Txn(ctx)

	clientv3.OpDelete(key)


	_, err := txn.Commit()
	return err
}

// GetRecursiveValue 获取前缀下的所有key
func (c *Etcd3Client) GetRecursiveValue(key string) (list []*Node, err error) {
	list = make([]*Node, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := c.Client.KV.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return nil, err
	}
	for _, kv := range resp.Kvs {
		list = append(list, &Node{
			Value:   string(kv.Value),
			FullDir: string(kv.Key),
			Version: kv.Version,
		})
	}

	return
}
