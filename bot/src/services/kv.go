package services

import "time"

var KV = KVStore{}

type KVStore map[string]*KValue

type KValue struct {
	Value  any
	Expiry time.Time
}

func (kv *KVStore) Get(key string) *KValue {
	println("Fetching key: " + key)
	if v, ok := (*kv)[key]; ok {
		if v.Expiry.After(time.Now()) {
			return v
		}
	}

	return nil
}

func (kv *KVStore) Set(key string, value any, expiry time.Duration) *KValue {
	v := &KValue{
		Value:  value,
		Expiry: time.Now().Add(expiry),
	}

	(*kv)[key] = v
	return v
}

func (kv *KVStore) Delete(key string) {
	delete((*kv), key)
}

func (kv *KVStore) Has(key string) bool {
	_, ok := (*kv)[key]
	return ok
}

func (kv *KVStore) Clear() {
	*kv = make(KVStore)
}

func (kv *KVStore) Expire(key string, duration time.Duration) {
	(*kv)[key].Expiry = time.Now().Add(duration)
}

func (kv *KVStore) ExpireAt(key string, expiry time.Time) {
	(*kv)[key].Expiry = expiry
}

func (kv *KVStore) GetExpiry(key string) time.Time {
	if v, ok := (*kv)[key]; ok {
		return v.Expiry
	}
	return time.Time{}
}

func (kv *KVStore) SetExpiry(key string, expiry time.Time) {
	(*kv)[key].Expiry = expiry
}

func (kv *KVStore) IsExpired(key string) bool {
	if v, ok := (*kv)[key]; ok {
		return v.Expiry.Before(time.Now())
	}
	return true
}

func (kv *KVStore) IsExpiredAt(key string, expiry time.Time) bool {
	if v, ok := (*kv)[key]; ok {
		return v.Expiry.Before(expiry)
	}
	return true
}
