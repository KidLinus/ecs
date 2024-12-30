package ecs

func (storage *Storage[ID]) Remove(id ID) {
	storage.lock.Lock()
	defer storage.lock.Unlock()
	entity, ok := storage.Entitys[id]
	if !ok {
		return
	}
	delete(storage.Entitys, id)
	storage.Compounds[entity.Compound].EntitysRemoved = sliceInsertOrdered(storage.Compounds[entity.Compound].EntitysRemoved, id)
}
