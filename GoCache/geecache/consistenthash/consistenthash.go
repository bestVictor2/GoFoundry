package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func([]byte) uint32

type Map struct { // 存储真实节点 & 虚拟节点
	hash     Hash
	replaces int
	keys     []int
	hashMap  map[int]string
}

func New(replicas int, hash Hash) *Map { // 初始化函数
	m := &Map{
		hash:     hash,
		replaces: replicas,
		hashMap:  make(map[int]string),
	}
	if hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) { // 传入真实节点，Hash映射虚拟节点
	for _, key := range keys {
		for i := 0; i < m.replaces; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			//fmt.Println(hash)
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string { // 通过Hash得到某值
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
