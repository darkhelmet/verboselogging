package post

type NotFound string

func (n NotFound) Error() string {
    return string(n)
}
