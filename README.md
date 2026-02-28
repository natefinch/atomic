你好！
很冒昧用这样的方式来和你沟通，如有打扰请忽略我的提交哈。我是光年实验室（gnlab.com）的HR，在招Golang开发工程师，我们是一个技术型团队，技术氛围非常好。全职和兼职都可以，不过最好是全职，工作地点杭州。
我们公司是做流量增长的，Golang负责开发SAAS平台的应用，我们做的很多应用是全新的，工作非常有挑战也很有意思，是国内很多大厂的顾问。
如果有兴趣的话加我微信：13515810775  ，也可以访问 https://gnlab.com/，联系客服转发给HR。
# atomic
    import "github.com/natefinch/atomic"
atomic is a go package for atomic file writing

By default, writing to a file in go (and generally any language) can fail
partway through... you then have a partially written file, which probably was
truncated when the write began, and bam, now you've lost data.

This go package avoids this problem, by writing first to a temp file, and then
overwriting the target file in an atomic way.  This is easy on linux, os.Rename
just is atomic.  However, on Windows, os.Rename is not atomic, and so bad things
can happen.  By wrapping the windows API moveFileEx, we can ensure that the move
is atomic, and we can be safe in knowing that either the move succeeds entirely,
or neither file will be modified.


## func ReplaceFile
``` go
func ReplaceFile(source, destination string) error
```
ReplaceFile atomically replaces the destination file or directory with the
source.  It is guaranteed to either replace the target file entirely, or not
change either file.


## func WriteFile
``` go
func WriteFile(filename string, r io.Reader) (err error)
```
WriteFile atomically writes the contents of r to the specified filepath.  If
an error occurs, the target file is guaranteed to be either fully written, or
not written at all.  WriteFile overwrites any file that exists at the
location (but only if the write fully succeeds, otherwise the existing file
is unmodified).

