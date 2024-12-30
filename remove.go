package ecs

func (storage *Storage[ID]) Remove(id ID) {
	entity, ok := storage.Entitys[id]
	if !ok {
		return
	}
	delete(storage.Entitys, id)
	storage.Compounds[entity.Compound].Removed = sliceInsertOrdered(storage.Compounds[entity.Compound].Removed, id)
}
