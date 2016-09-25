package gls

func Go(f func()) {
	_go(f)
}

func Put(key, value interface{}) {
	put(key, value)
}

func Get(key interface{}) (value interface{}, ok bool) {
	return get(key)
}

func Delete(key interface{}) {
	del(key)
}
