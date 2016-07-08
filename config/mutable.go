// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"errors"

	pbconfig "github.com/maniksurtani/quotaservice/protos/config"
)

func CreateBucket(clonedCfg *pbconfig.ServiceConfig, namespace string, b *pbconfig.BucketConfig) error {
	if namespace == GlobalNamespace {
		if clonedCfg.GlobalDefaultBucket != nil {
			return errors.New("GlobalDefaultBucket already exists")
		}

		clonedCfg.GlobalDefaultBucket = b
	} else {
		ns := clonedCfg.Namespaces[namespace]

		if ns == nil {
			return errors.New("Namespace doesn't exist")
		}

		if b.Name == DefaultBucketName {
			if ns.DefaultBucket != nil {
				return errors.New("DefaultBucket already exists")
			}

			ns.DefaultBucket = b
		} else if b.Name == DynamicBucketTemplateName {
			if ns.DynamicBucketTemplate != nil {
				return errors.New("DynamicBucketTemplate already exists")
			}

			ns.DynamicBucketTemplate = b
		} else if ns.Buckets[b.Name] != nil {
			return errors.New("Bucket " + b.Name + " already exists")
		} else {
			ns.Buckets[b.Name] = b
		}
	}

	return nil
}

func UpdateBucket(clonedCfg *pbconfig.ServiceConfig, namespace string, b *pbconfig.BucketConfig) error {
	if namespace == GlobalNamespace {
		clonedCfg.GlobalDefaultBucket = b
	} else {
		ns := clonedCfg.Namespaces[namespace]

		if ns == nil {
			return errors.New("Namespace doesn't exist")
		}

		if b.Name == DefaultBucketName {
			ns.DefaultBucket = b
		} else if b.Name == DynamicBucketTemplateName {
			ns.DynamicBucketTemplate = b
		} else {
			ns.Buckets[b.Name] = b
		}
	}

	return nil
}

func DeleteBucket(clonedCfg *pbconfig.ServiceConfig, namespace, name string) error {
	if namespace == GlobalNamespace {
		clonedCfg.GlobalDefaultBucket = nil
	} else {
		ns := clonedCfg.Namespaces[namespace]

		if ns == nil {
			return errors.New("No such namespace " + namespace + ".")
		}

		if name == DefaultBucketName {
			ns.DefaultBucket = nil
		} else if name == DynamicBucketTemplateName {
			ns.DynamicBucketTemplate = nil
		} else {
			delete(ns.Buckets, name)
		}
	}

	return nil
}

func DeleteNamespace(clonedCfg *pbconfig.ServiceConfig, n string) error {
	if clonedCfg.Namespaces[n] == nil {
		return errors.New("No such namespace " + n)
	}

	delete(clonedCfg.Namespaces, n)

	return nil
}

func CreateNamespace(clonedCfg *pbconfig.ServiceConfig, nsCfg *pbconfig.NamespaceConfig) error {
	if clonedCfg.Namespaces[nsCfg.Name] != nil {
		return errors.New("Namespace " + nsCfg.Name + " already exists.")
	}

	return UpdateNamespace(clonedCfg, nsCfg)
}

func UpdateNamespace(clonedCfg *pbconfig.ServiceConfig, nsCfg *pbconfig.NamespaceConfig) error {
	if clonedCfg.Namespaces == nil {
		clonedCfg.Namespaces = make(map[string]*pbconfig.NamespaceConfig)
	}

	clonedCfg.Namespaces[nsCfg.Name] = nsCfg

	return nil
}
