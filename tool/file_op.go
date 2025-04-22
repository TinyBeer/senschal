package tool

import (
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
)

func Copy(srcPath, dstPath string) error {
	src, err := newPath(srcPath)
	if err != nil {
		return err
	}
	err = src.syncStat()
	if err != nil {
		return err
	}
	dst, err := newPath(dstPath)
	if err != nil {
		return err
	}
	srcIsDir, err := src.IsDir()
	if err != nil {
		return err
	}
	exist, err := dst.IsExist()
	if err != nil {
		return err
	}
	dstIsDir := false
	if !exist {
		dstIsDir = srcIsDir
	} else {
		dstIsDir, err = dst.IsDir()
		if err != nil {
			return err
		}
	}
	if srcIsDir && !dstIsDir {
		return errors.New("can not copy a dir to a file")
	}
	if !srcIsDir {
		log.Printf("copy %s to %s\n", srcPath, dstPath)
		if !dstIsDir {
			return cpFile(srcPath, dstPath)
		} else {
			return cpFile(srcPath, filepath.Join(dstPath, filepath.Base(srcPath)))
		}
	}
	for _, child := range src.Children {
		err := copy(filepath.Join(child.BasePath, child.RelativePath, child.Name), filepath.Join(dstPath, filepath.Base(srcPath), child.RelativePath, child.Name))
		if err != nil {
			return err
		}
	}
	return nil
}

func copy(srcPath, dstPath string) error {
	src, err := newPath(srcPath)
	if err != nil {
		return err
	}
	err = src.syncStat()
	if err != nil {
		return err
	}
	dst, err := newPath(dstPath)
	if err != nil {
		return err
	}
	srcIsDir, err := src.IsDir()
	if err != nil {
		return err
	}
	exist, err := dst.IsExist()
	if err != nil {
		return err
	}
	dstIsDir := false
	if !exist {
		dstIsDir = srcIsDir
	} else {
		dstIsDir, err = dst.IsDir()
		if err != nil {
			return err
		}
	}
	if srcIsDir && !dstIsDir {
		return errors.New("can not copy a dir to a file")
	}
	if !srcIsDir {
		log.Printf("copy %s to %s\n", srcPath, dstPath)
		if !dstIsDir {
			return cpFile(srcPath, dstPath)
		} else {
			return cpFile(srcPath, filepath.Join(dstPath, filepath.Base(srcPath)))
		}
	}
	for _, child := range src.Children {
		err := copy(filepath.Join(child.BasePath, child.RelativePath, child.Name), filepath.Join(dstPath, child.RelativePath, child.Name))
		if err != nil {
			return err
		}
	}
	return nil
}

func cpFile(srcPath, dstPath string) error {
	srcFile, err := NewFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to new source file with arg %v, err:%v", srcPath, err)
	}
	srcSize, err := srcFile.GetSize()
	if err != nil {
		return fmt.Errorf("failed to get source file size with arg %v, err:%v", dstPath, err)
	}

	dstFile, err := NewFile(dstPath)
	if err != nil {
		return fmt.Errorf("failed to new destination file with arg %v, err:%v", dstPath, err)
	}
	dst, err := dstFile.GetWriter()
	if err != nil {
		return fmt.Errorf("failed to get writer from dst file, err:%v", err)
	}

	src, err := srcFile.GetReader()
	if err != nil {
		return fmt.Errorf("failed to get reader from src file, err:%v", err)
	}
	byteCnt, err := io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy file from %v to %v, err:%v", srcPath, dstPath, err)
	}
	log.Printf("copy %v byte from %v to %v", byteCnt, srcPath, dstPath)

	dstSize, err := dstFile.GetSize()
	if err != nil {
		return fmt.Errorf("failed to get destination file size with arg %v, err:%v", dstPath, err)
	}

	if srcSize != byteCnt || srcSize != dstSize {
		return errors.New("failed to copy file")
	}

	return nil
}
