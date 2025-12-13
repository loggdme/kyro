package kyro

import "os"

func SafeRemoveFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return err
		}
		return err
	}

	return nil
}
