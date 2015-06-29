package config

//
// func ParseConfigFile(buff []byte) (*AppConfig, error) {
// 	var err error
// 	ac := NewDefaultConfig()
//
// 	// this will be used to hash all the files thar are opened while parsing
// 	hasher := md5.New()
// 	hasher.Write(buff)
//
// 	err = json.Unmarshal(buff, ac)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	ac.KeepAliveAge, err = time.ParseDuration(ac.Raw_KeepAliveAge)
// 	if err != nil {
// 		return ac, err
// 	}
//
// 	paths, err := filepath.Glob(ac.EscalationsDir + "*.json")
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	for _, path := range paths {
// 		buff, err := loadFile(path)
// 		if err != nil {
// 			return ac, err
// 		}
//
// 		hasher.Write(buff)
// 		p, err := loadPolicy(buff)
// 		if err != nil {
// 			return ac, err
// 		}
//
// 		// set up the file name for the policy
// 		if p.Name == "" {
// 			path = filepath.Base(path)
// 			p.Name = path[:len(path)-4]
// 		}
//
// 		ac.Policies = append(ac.Policies, p)
// 	}
//
// 	if ac.GlobalPolicy != nil {
// 		ac.GlobalPolicy.Compile()
// 	}
//
// 	if ac.EventProviders == nil {
// 		ac.EventProviders = &provider.EventProviderCollection{}
// 	}
//
// 	if ac.LogLevel == "" {
// 		ac.LogLevel = DefaultLogLevel
// 	}
//
// 	ac.Hash = hasher.Sum(nil)
//
// 	return ac, nil
//
// }
//
// func LoadConfigFile(fileName string) (*AppConfig, error) {
// 	buff, err := ioutil.ReadFile(fileName)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	ac, err := ParseConfigFile(buff)
// 	if err != nil {
// 		logrus.Error(err)
// 		return ac, err
// 	}
//
// 	ac.fileName = fileName
// 	return ac, err
//
// }
//
// func loadFile(fileName string) ([]byte, error) {
// 	if !filepath.IsAbs(fileName) {
// 		fileName, _ = filepath.Abs(fileName)
//
// 	}
// 	return ioutil.ReadFile(fileName)
// }
