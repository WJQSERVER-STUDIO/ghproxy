package gitclone

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/pierrec/lz4"
)

func CloneRepo(dir string, repoName string, repoUrl string) error {
	repoPath := dir
	_, err := git.PlainClone(repoPath, true, &git.CloneOptions{
		URL:      repoUrl,
		Progress: os.Stdout,
		Mirror:   true,
	})
	if err != nil && !errors.Is(err, git.ErrRepositoryAlreadyExists) {
		fmt.Printf("Fail to clone: %v\n", err)
	} else if err != nil && errors.Is(err, git.ErrRepositoryAlreadyExists) {
		// 移除文件夹
		fmt.Printf("Repository already exists\n")
		err = os.RemoveAll(repoPath)
		if err != nil {
			fmt.Printf("Fail to remove: %v\n", err)
			return err
		}
		_, err = git.PlainClone(repoPath, true, &git.CloneOptions{
			URL:      repoUrl,
			Progress: os.Stdout,
			Mirror:   true,
		})
		if err != nil {
			fmt.Printf("Fail to clone: %v\n", err)
			return err
		}
	}

	// 压缩
	err = CompressRepo(repoPath)
	if err != nil {
		fmt.Printf("Fail to compress: %v\n", err)
		return err
	}
	return nil
}

// CompressRepo 将指定的仓库压缩成 LZ4 格式的压缩包
func CompressRepo(repoPath string) error {
	lz4File, err := os.Create(repoPath + ".lz4")
	if err != nil {
		return fmt.Errorf("failed to create LZ4 file: %w", err)
	}
	defer lz4File.Close()

	// 创建 LZ4 编码器
	lz4Writer := lz4.NewWriter(lz4File)
	defer lz4Writer.Close()

	// 创建 tar.Writer
	tarBuffer := new(bytes.Buffer)
	tarWriter := tar.NewWriter(tarBuffer)

	// 遍历仓库目录并打包
	err = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 创建 tar 文件头
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name, err = filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}

		// 写入 tar 文件头
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// 如果是文件，写入文件内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk through repo directory: %w", err)
	}

	// 关闭 tar.Writer
	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	// 将 tar 数据写入 LZ4 压缩包
	if _, err := lz4Writer.Write(tarBuffer.Bytes()); err != nil {
		return fmt.Errorf("failed to write to LZ4 file: %w", err)
	}

	return nil
}
