package gson
import(
	"unsafe"
)
type PointerItem unsafe.Pointer
type PointStack struct{
	m_items []PointerItem
}
//push operation
func(s *PointStack) push(item PointerItem){
	s.m_items = append(s.m_items, item)
}
//pop operation, pop the specified number of items
func(s *PointStack) pop(nSize int) []PointerItem {
	nLen := len(s.m_items)
	if nSize > nLen{
		panic("stack size out of boundary")
	}
	res := s.m_items[nLen - 1 - nSize : nLen]
	s.m_items = s.m_items[0 : nLen - 1 - nSize]
	return res
}
//get the top item
func(s *PointStack) top() *PointerItem{
	nLen := len(s.m_items)
	if 0 == nLen{
		return nil
	}
	return &s.m_items[nLen - 1]
}