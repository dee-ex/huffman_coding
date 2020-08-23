package main

import "fmt"
import "bufio"
import "os"
import "strings"
import "strconv"
import "sort"
import "sync"
import "flag"

var wg sync.WaitGroup
var col_wg sync.WaitGroup
var lock sync.Mutex

type node struct {
    label string
    left string
    right string
    value int
}

func init_node(path string) []node {
    f, err := os.Open(path)
    if err != nil {
        return nil
    }

    defer f.Close()

    var nodes []node
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        data := strings.Split(scanner.Text(), " ")
        label := data[0]
        v64, _ := strconv.ParseInt(data[1], 10, 32)
        value := int(v64)
        nodes = append(nodes, node{label: label, value: value})
    }
    return nodes
}

func rerange_nodes(nodes []node) []node {
    sort.SliceStable(nodes, func(i, j int) bool {
        return nodes[i].value < nodes[j].value
    })
    return nodes
}

func find_node(label string, nodes []node) node {
    for _, n := range nodes {
        if n.label == label {
            return n
        }
    }
    return node{}
}

func from_root(n node, nodes []node,temp string, collect chan string) {
    defer wg.Done()

    left_node := find_node(n.left, nodes)
    right_node := find_node(n.right, nodes)

    if left_node.left == "" {
        collect <- left_node.label + ":" + temp + "1"
    } else {
        wg.Add(1)
        go from_root(left_node, nodes, temp + "1", collect)
    }

    if right_node.left == "" {
        collect <- right_node.label + ":" + temp + "0"
    } else {
        wg.Add(1)
        go from_root(right_node, nodes, temp + "0", collect)
    }
}

func build_tree(nodes []node) []node {
    wk_nodes := make([]node, len(nodes))
    copy(wk_nodes, nodes)
    var counter int

    for ;len(wk_nodes) > 1; {
        counter++
        wk_nodes = rerange_nodes(wk_nodes)
        new_label := fmt.Sprintf("node%d", counter)
        new_value := wk_nodes[0].value + wk_nodes[1].value
        left := wk_nodes[0].label
        right := wk_nodes[1].label
        new_node := node{label: new_label, value: new_value,
                        left: left, right: right}
        wk_nodes = wk_nodes[2:]
        nodes = append(nodes, new_node)
        wk_nodes = append([]node{new_node}, wk_nodes...)
    }
    return nodes
}

func sort_by_len(arr []string) []string {
    sort.SliceStable(arr, func(i, j int) bool {
        return len(arr[i]) > len(arr[j])
    })
    return arr
}

func main() {
    var path string
    flag.StringVar(&path, "p", "", "input's path")
    flag.Parse()
    if path == "" {
        fmt.Println("run -h to see more information")
        return
    }

    nodes := init_node(path)
    collect := make(chan string, len(nodes))
    nodes = build_tree(nodes)

    last_node := nodes[len(nodes) - 1]
    wg.Add(1)
    go func(last_node node, nodes []node) {
        from_root(last_node, nodes, "", collect)
        wg.Wait()
        close(collect)
    }(last_node, nodes)

    var res []string
    for item := range collect {
        col_wg.Add(1)
        go func(item string) {
            defer col_wg.Done()
            lock.Lock()
            res = append(res, item)
            lock.Unlock()
        }(item)
    }
    col_wg.Wait()

    res = sort_by_len(res)
    for _, item := range res {
        fmt.Println(item)
    }
}
