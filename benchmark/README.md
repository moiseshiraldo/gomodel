# Introduction

Each benchmark compares the performance of a specific database operation for the following APIs:

- **Raw SQL**: plain SQL using the [database/sql](https://golang.org/pkg/database/sql/) pacakge.
- **GoModel (map container)**: gomodel with the default [Values](https://godoc.org/github.com/moiseshiraldo/gomodel#Values) type to store the information.
- **GoModel (struct container)**: gomodel with a struct type to store the information.
- **GoModel (builder container)**: gomodel with a struct type implementing the [Builder](https://godoc.org/github.com/moiseshiraldo/gomodel#Builder) interface to store the information.

# Operations

The basic operations being compared are the following:

- **Insert**: inserting values to a database table.
- **Read**: getting a specific row from a database table.
- **MultiRead**: getting 100 rows from a database table.
- **Update**: updating 100 rows on a database table.

# Results

## Environment

The results were obtained using an in-memory SQLite database on a machine with the following details:

- **CPU**: 3.5GHz AMD FX-6300 six-core processor.
- **RAM**: 8GB 800MHz DDR3.
- **OS**: Slackware GNU/Linux (kernel 4.19.46).
- **Go**: go1.12.2 gccgo (GCC) 9.1.0 linux/amd64.

## Reports

### Insert
| API               | Time (ns/op) | Memory usage (B/op) | Memory allocation (allocs/op) |
|-------------------|:------------:|:-------------------:|:-----------------------------:|
| Raw SQL           | 79940        | 816                 | 27                            |
| GoModel (map)     | 114343       | 3644                | 107                           |
| GoModel (struct)  | 118976       | 3392                | 105                           |
| GoModel (builder) | 112465       | 3395                | 105                           |

### Read
| API               | Time (ns/op) | Memory usage (B/op) | Memory allocation (allocs/op) |
|-------------------|:------------:|:-------------------:|:-----------------------------:|
| Raw SQL           | 153949       | 1808                | 69                            |
| GoModel (map)     | 253501       | 4645                | 174                           |
| GoModel (struct)  | 248816       | 3731                | 142                           |
| GoModel (builder) | 257641       | 4002                | 155                           |

### MultiRead
| API               | Time (ns/op) | Memory usage (B/op) | Memory allocation (allocs/op) |
|-------------------|:------------:|:-------------------:|:-----------------------------:|
| Raw SQL           | 5523112      | 70429               | 3246                          |
| GoModel (map)     | 6656044      | 156256              | 7412                          |
| GoModel (struct)  | 6165835      | 109878              | 5313                          |
| GoModel (builder) | 6132829      | 113084              | 5709                          |

### Update
| API               | Time (ns/op) | Memory usage (B/op) | Memory allocation (allocs/op) |
|-------------------|:------------:|:-------------------:|:-----------------------------:|
| Raw SQL           | 77102        | 570                 | 20                            |
| GoModel (map)     | 101985       | 3329                | 82                            |
| GoModel (struct)  | 141747       | 4219                | 118                           |
| GoModel (builder) | 127645       | 4329                | 121                           |
