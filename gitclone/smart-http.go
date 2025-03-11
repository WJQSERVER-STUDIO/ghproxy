package gitclone

/*
package gitclone

import (
	"compress/gzip"
	"ghproxy/config"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
)

// MIT https://github.com/erred/gitreposerver

// httpInfoRefs 函数处理 /info/refs 请求，用于 Git 客户端获取仓库的引用信息。
// 返回一个 gin.HandlerFunc 类型的处理函数。
func HttpInfoRefs(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		repo := c.Param("repo") // 从 Gin 上下文中获取路由参数 "repo"，即仓库名
		username := c.Param("username")
		repoName := repo
		dir := cfg.GitClone.Dir + "/" + username + "/" + repo
		url := "https://github.com/" + username + "/" + repo

		// 输出 repo user dir url
		logInfo("Repo: %s, User: %s, Dir: %s, Url: %s\n", repoName, username, dir, url)

		_, err := os.Stat(dir) // 检查目录是否存在
		if os.IsNotExist(err) {
			CloneRepo(dir, repoName, url)
		}

		// 检查请求参数 "service" 是否为 "git-upload-pack"。
		// 这是为了确保只处理 smart git 的 upload-pack 服务请求。
		if c.Query("service") != "git-upload-pack" {
			c.String(http.StatusForbidden, "only smart git")                                                       // 如果 service 参数不正确，返回 403 Forbidden 状态码和错误信息
			log.Printf("Request to /info/refs with invalid service: %s, repo: %s\n", c.Query("service"), repoName) // 记录无效 service 参数的日志
			return                                                                                                 // 结束处理
		}

		c.Header("content-type", "application/x-git-upload-pack-advertisement") // 设置 HTTP 响应头的 Content-Type 为 advertisement 类型。
		// 这种类型用于告知客户端服务器支持的 Git 服务。

		ep, err := transport.NewEndpoint("/") // 创建一个新的传输端点 (Endpoint)。这里使用根路径 "/" 作为端点，表示本地文件系统。
		if err != nil {                       // 检查创建端点是否出错
			log.Printf("Error creating endpoint: %v, repo: %s\n", err, repoName) // 记录创建端点错误日志
			c.String(http.StatusInternalServerError, err.Error())                // 返回 500 Internal Server Error 状态码和错误信息
			return                                                               // 结束处理
		}

		bfs := osfs.New(dir)                           // 创建一个基于本地文件系统的 billy Filesystem (bfs)。dir 变量指定了仓库的根目录。
		ld := server.NewFilesystemLoader(bfs)          // 创建一个基于文件系统的仓库加载器 (Loader)。Loader 负责从文件系统中加载仓库。
		svr := server.NewServer(ld)                    // 创建一个新的 Git 服务器 (Server)。Server 负责处理 Git 服务请求。
		sess, err := svr.NewUploadPackSession(ep, nil) // 创建一个新的 upload-pack 会话 (Session)。Session 用于处理客户端的 upload-pack 请求。
		if err != nil {                                // 检查创建会话是否出错
			log.Printf("Error creating upload pack session: %v, repo: %s\n", err, repoName) // 记录创建会话错误日志
			c.String(http.StatusInternalServerError, err.Error())                           // 返回 500 Internal Server Error 状态码和错误信息
			return                                                                          // 结束处理
		}

		ar, err := sess.AdvertisedReferencesContext(c.Request.Context()) // 获取已通告的引用 (Advertised References)。Advertised References 包含了仓库的分支、标签等信息。
		if err != nil {                                                  // 检查获取 Advertised References 是否出错
			c.String(http.StatusInternalServerError, err.Error())                            // 返回 500 Internal Server Error 状态码和错误信息
			log.Printf("Error getting advertised references: %v, repo: %s\n", err, repoName) // 记录获取 Advertised References 错误日志
			return                                                                           // 结束处理
		}

		// 设置 Advertised References 的前缀 (Prefix)。
		// Prefix 通常包含 # service=git-upload-pack 和 pktline.Flush。
		// # service=git-upload-pack 用于告知客户端服务器提供的是 upload-pack 服务。
		// pktline.Flush 用于在 pkt-line 格式中发送 flush-pkt。
		ar.Prefix = [][]byte{
			[]byte("# service=git-upload-pack"), // 服务类型声明
			pktline.Flush,                       // pkt-line flush 信号
		}
		err = ar.Encode(c.Writer) // 将 Advertised References 编码并写入 HTTP 响应。使用 pkt-line 格式进行编码。
		if err != nil {           // 检查编码和写入是否出错
			log.Printf("Error encoding advertised references: %v, repo: %s\n", err, repoName) // 记录编码错误日志
			c.String(http.StatusInternalServerError, err.Error())                             // 返回 500 Internal Server Error 状态码和错误信息
			return                                                                            // 结束处理
		}
	}
}

// httpGitUploadPack 函数处理 /git-upload-pack 请求，用于处理 Git 客户端的推送 (push) 操作。
// 返回一个 gin.HandlerFunc 类型的处理函数。
func HttpGitUploadPack(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		repo := c.Param("repo") // 从 Gin 上下文中获取路由参数 "repo"，即仓库名
		username := c.Param("username")
		repoName := repo
		dir := cfg.GitClone.Dir + "/" + username + "/" + repo

		c.Header("content-type", "application/x-git-upload-pack-result") // 设置 HTTP 响应头的 Content-Type 为 result 类型。
		// 这种类型用于返回 upload-pack 操作的结果。

		var bodyReader io.Reader = c.Request.Body // 初始化 bodyReader 为 HTTP 请求的 body。用于读取客户端发送的数据。
		// 检查请求头 "Content-Encoding" 是否为 "gzip"。
		// 如果是 gzip，则需要使用 gzip 解压缩请求 body。
		if c.GetHeader("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(c.Request.Body) // 创建一个新的 gzip Reader，用于解压缩请求 body。
			if err != nil {                                   // 检查创建 gzip Reader 是否出错
				log.Printf("Error creating gzip reader: %v, repo: %s\n", err, repoName) // 记录创建 gzip Reader 错误日志
				c.String(http.StatusInternalServerError, err.Error())                   // 返回 500 Internal Server Error 状态码和错误信息
				return                                                                  // 结束处理
			}
			defer gzipReader.Close() // 延迟关闭 gzip Reader，确保资源释放
			bodyReader = gzipReader  // 将 bodyReader 替换为 gzip Reader，后续从 gzip Reader 中读取数据
		}

		upr := packp.NewUploadPackRequest() // 创建一个新的 UploadPackRequest 对象。UploadPackRequest 用于解码客户端发送的 upload-pack 请求数据。
		err := upr.Decode(bodyReader)       // 解码请求 body 中的数据到 UploadPackRequest 对象中。使用 packp 协议格式进行解码。
		if err != nil {                     // 检查解码是否出错
			log.Printf("Error decoding upload pack request: %v, repo: %s\n", err, repoName) // 记录解码错误日志
			c.String(http.StatusInternalServerError, err.Error())                           // 返回 500 Internal Server Error 状态码和错误信息
			return                                                                          // 结束处理
		}

		ep, err := transport.NewEndpoint("/") // 创建一个新的传输端点 (Endpoint)。这里使用根路径 "/" 作为端点，表示本地文件系统。
		if err != nil {                       // 检查创建端点是否出错
			log.Printf("Error creating endpoint: %v, repo: %s\n", err, repoName) // 记录创建端点错误日志
			c.String(http.StatusInternalServerError, err.Error())                // 返回 500 Internal Server Error 状态码和错误信息
			return                                                               // 结束处理
		}

		bfs := osfs.New(dir)                           // 创建一个基于本地文件系统的 billy Filesystem (bfs)。dir 变量指定了仓库的根目录。
		ld := server.NewFilesystemLoader(bfs)          // 创建一个基于文件系统的仓库加载器 (Loader)。Loader 负责从文件系统中加载仓库。
		svr := server.NewServer(ld)                    // 创建一个新的 Git 服务器 (Server)。Server 负责处理 Git 服务请求。
		sess, err := svr.NewUploadPackSession(ep, nil) // 创建一个新的 upload-pack 会话 (Session)。Session 用于处理客户端的 upload-pack 请求。
		if err != nil {                                // 检查创建会话是否出错
			log.Printf("Error creating upload pack session: %v, repo: %s\n", err, repoName) // 记录创建会话错误日志
			c.String(http.StatusInternalServerError, err.Error())                           // 返回 500 Internal Server Error 状态码和错误信息
			return                                                                          // 结束处理
		}

		res, err := sess.UploadPack(c.Request.Context(), upr) // 处理 upload-pack 请求，执行实际的仓库推送操作。
		// sess.UploadPack 函数接收 context 和 UploadPackRequest 对象作为参数，返回 UploadPackResult 和 error。
		if err != nil { // 检查 UploadPack 操作是否出错
			c.String(http.StatusInternalServerError, err.Error())                 // 返回 500 Internal Server Error 状态码和错误信息
			log.Printf("Error during upload pack: %v, repo: %s\n", err, repoName) // 记录 UploadPack 操作错误日志
			return                                                                // 结束处理
		}

		err = res.Encode(c.Writer) // 将 UploadPackResult 编码并写入 HTTP 响应。使用 pkt-line 格式进行编码。
		if err != nil {            // 检查编码和写入是否出错
			log.Printf("Error encoding upload pack result: %v, repo: %s\n", err, repoName) // 记录编码错误日志
			c.String(http.StatusInternalServerError, err.Error())                          // 返回 500 Internal Server Error 状态码和错误信息
			return                                                                         // 结束处理
		}
	}
}

*/
