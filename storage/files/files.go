package files

import (
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"tg-bot/lib/e"
	"tg-bot/storage"
)

type Storage struct {
	basePath string
}

const defaultPerm = 0774

func New(basePath string) Storage {
	return Storage{basePath: basePath}
}

func (s Storage) Save(page *storage.Page) (err error) {
	defer func() { err = e.WrapIfErr("can't save page", err) }()

	// Формирует путь, где будет лежать файл
	fPath := filepath.Join(s.basePath, page.UserName)

	// Создать директории по этому пути
	if err := os.MkdirAll(fPath, defaultPerm); err != nil {
		return err
	}

	// Создаст имя файла
	fName, err := fileName(page)
	if err != nil {
		return err
	}

	// Путь с именем файла
	fPath = filepath.Join(fPath, fName)

	file, err := os.Create(fPath)
	if err != nil {
		return err
	}

	// Без такой конструкции (функции-обёртки) было бы предупреждение о необработанной ошибке
	defer func() { _ = file.Close() }()

	// gob - Пакет для сериализации и десериализации данных в бинарный формат
	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return err
	}

	return nil
}

func (s Storage) PickRandom(userName string) (page *storage.Page, err error) {
	defer func() { err = e.WrapIfErr("can't save page", err) }()

	path := filepath.Join(s.basePath, userName)

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	rand.Seed(time.Now().UnixNano()) // Генерация rand на основе метки времени
	n := rand.Intn(len(files))       // Указываем верхнюю границу рандома (не больше кол-ва всех файлов)

	file := files[n]

	return s.decodePage(filepath.Join(path, file.Name()))
}

func (s Storage) Remove(p *storage.Page) error {
	fileName, err := fileName(p)
	if err != nil {
		return e.Wrap("can't remove file", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	if err := os.Remove(path); err != nil {
		msg := fmt.Sprintf("can't remove file %s", path)
		return e.Wrap(msg, err)
	}

	return nil
}

func (s Storage) IsExists(p *storage.Page) (bool, error) {
	fileName, err := fileName(p)
	if err != nil {
		return false, e.Wrap("can't check if file exists", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	// os.Stat - возвращает FileInfo, описывающий именованный файл
	switch _, err = os.Stat(path); {
	case errors.Is(err, os.ErrNotExist): // Если файл не был найден
		return false, nil
	case err != nil: // Описывает любые другие ошибки
		msg := fmt.Sprintf("can't check if file %s exists", path)

		return false, e.Wrap(msg, err)
	}

	return true, nil
}

func (s Storage) decodePage(filePath string) (*storage.Page, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, e.Wrap("can't decode page", err)
	}
	defer func() { _ = f.Close() }()

	var p storage.Page

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	return &p, nil
}

// fileName - вернёт имя файла (Хэш) на основе его имени и пути
func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}
