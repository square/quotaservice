// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"errors"

	pbconfig "github.com/maniksurtani/quotaservice/protos/config"
)

func CreateBucket(clonedCfg *pbconfig.ServiceConfig, namespace string, b *pbconfig.BucketConfig) error {
	if namespace == GlobalNamespace {
		// TODO(steved) what to do?
	} else {
		ns := clonedCfg.Namespaces[namespace]

		if ns == nil {
			return errors.New("Namespace doesn't exist")
		}

		if ns.Buckets[b.Name] != nil {
			return errors.New("Bucket " + b.Name + " already exists")
		} else {
			ns.Buckets[b.Name] = b
		}
	}

	return nil
}

func UpdateBucket(clonedCfg *pbconfig.ServiceConfig, namespace string, b *pbconfig.BucketConfig) error {
	if namespace == GlobalNamespace {
		// TODO(steved) what to do?
	} else {
		ns := clonedCfg.Namespaces[namespace]

		if ns == nil {
			return errors.New("Namespace doesn't exist")
		}

		ns.Buckets[b.Name] = b
	}

	return nil
}

func DeleteBucket(clonedCfg *pbconfig.ServiceConfig, namespace, name string) error {
	if namespace == GlobalNamespace {
		// TODO(steved) confirm this whole condition, it's weird
	} else {
		ns := clonedCfg.Namespaces[namespace]

		if ns == nil {
			return errors.New("No such namespace " + namespace + ".")
		}

		// TODO(steved) should this be applied to all Bucket ops?
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
