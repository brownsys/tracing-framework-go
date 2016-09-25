package runtime

// TODO(jliebowf): Document

func GetLocal() interface{} {
	return getg().local
}

func SetLocal(local interface{}) {
	getg().local = local
}
