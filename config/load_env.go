package config

func loadEnv(path string) {
        f, err := os.Open(path)
        if err != nil {
                return
        }
        defer f.Close()

        sc := bufio.NewScanner(f)
        for sc.Scan() {
                line := strings.TrimSpace(sc.Text())
                if line == "" || strings.HasPrefix(line, "#") {
                        continue
                }
                key, val, ok := strings.Cut(line, "=")
                if !ok {
                        continue
                }
                key = strings.TrimSpace(key)
                val = strings.Trim(strings.TrimSpace(val), `"'`)
                if _, exists := os.LookupEnv(key); !exists {
                        os.Setenv(key, val)
                }
        }
}