package main

type ListNode struct {
	Val  int       // 节点值
	Next *ListNode // 指向下一个节点的指针
}

func NewListNode(val int) *ListNode {
	return &ListNode{
		Val:  val,
		Next: nil,
	}
}

func main() {
	// 初始化链表
	n0 := NewListNode(1)
	n1 := NewListNode(3)
	n2 := NewListNode(2)
	n3 := NewListNode(5)
	n4 := NewListNode(4)

	// 构建引用指向
	n0.Next = n1
	n1.Next = n2
	n2.Next = n3
	n3.Next = n4
}

/* 在链表的节点 n0 之后插入节点 P */
func insertNode(n0 *ListNode, P *ListNode) {
	n1 := n0.Next
	P.Next = n1
	n0.Next = P
}

/* 在链表中删除节点 P */
func removeNode(n0 *ListNode) {
	if n0.Next == nil {
		return
	}
	// n0 -> P -> n1
	P := n0.Next
	n1 := P.Next
	n0.Next = n1
}

/* 访问链表中索引为 index 的节点 */
func access(head *ListNode, index int) *ListNode {
	for i := 0; i < index; i++ {
		if head == nil {
			return nil
		}
		head = head.Next
	}
	return head
}

/* 在链表中查找值为 target 的首个节点 */
func findNode(head *ListNode, target int) int {
	var index int
	for head != nil {
		if head.Val == target {
			return index
		}
		head = head.Next
		index++
	}
	return -1
}
