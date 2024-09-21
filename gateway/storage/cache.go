package storage

type Node struct {
	Key  string
	Val  bool
	Prev *Node
	Next *Node
}

type Cache struct {
	Cache    map[string]*Node
	First    *Node
	Last     *Node
	Capacity int
}

func NewCache(capacity int) *Cache {
	cache := &Cache{
		Cache:    make(map[string]*Node),
		Capacity: capacity,
	}
	cache.First = &Node{"", true, nil, nil}
	cache.Last = &Node{"", true, nil, nil}
	cache.First.Prev = cache.Last
	cache.Last.Next = cache.First
	return cache
}

func (cache *Cache) Remove(node *Node) {
	prev := node.Prev
	next := node.Next
	prev.Next = next
	next.Prev = prev
}

func (cache *Cache) Insert(node *Node) {
	node.Prev = cache.First.Prev
	node.Next = cache.First
	cache.First.Prev.Next = node
	cache.First.Prev = node
}

func (cache *Cache) Get(key string) bool {
	if node, exists := cache.Cache[key]; exists {
		cache.Remove(node)
		cache.Insert(node)
		return node.Val
	} else {
		return false
	}
}

func (cache *Cache) Put(key string) string {
	if node, exists := cache.Cache[key]; exists {
		cache.Remove(node)
	}
	newNode := &Node{Key: key, Val: true}
	cache.Cache[key] = newNode
	cache.Insert(newNode)

	if len(cache.Cache) > cache.Capacity {
		lru := cache.Last.Next
		cache.Remove(lru)
		delete(cache.Cache, lru.Key)
	}
	return newNode.Key
}
