/*
 * Copyright 2018 - 2019 Christian Müller <dev@c-mueller.xyz>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ads

import (
	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"time"
)

func init() {
	caddy.RegisterPlugin("ads", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	log.Info("Launching CoreDNS Ads Plugin")
	c.Next()
	cfg, err := parsePluginConfiguration(c)
	if err != nil {
		return err
	}

	updater := &ListUpdater{
		Enabled:         cfg.EnableAutoUpdate,
		RetryCount:      cfg.ListRenewalRetryCount,
		RetryDelay:      cfg.ListRenewalRetryInterval,
		UpdateInterval:  cfg.HttpListRenewalInterval,
		Plugin:          nil,
		persistLists:    cfg.EnableListPersistence,
		persistencePath: cfg.ListPersistencePath,
	}

	c.OnStartup(func() error {
		once.Do(func() {
			metrics.MustRegister(c, requestCount)
			metrics.MustRegister(c, blockedRequestCount)
			metrics.MustRegister(c, requestCountBySource)
			metrics.MustRegister(c, blockedRequestCountBySource)
			updater.Start()
		})
		return nil
	})

	blockageMap := make(ListMap, 0)

	ruleset, err := buildRulesetFromConfig(cfg)
	if err != nil {
		return err
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {

		adsPlugin := DNSAdBlock{
			Next:       next,
			Blacklists: cfg.BlacklistURLs,
			blacklist:  blockageMap,
			RuleSet:    *ruleset,
			config:     cfg,
		}

		updater.Plugin = &adsPlugin

		if !cfg.EnableAutoUpdate {
			adsPlugin.updater = updater
		}

		return &adsPlugin
	})

	return nil
}

func persistLoadedBlocklist(updater *ListUpdater, enableAutoUpdate bool, blocklists []string, blockageMap ListMap, persistedBlocklistPath string) {
	updater.lastPersistenceUpdate = time.Now()
	if enableAutoUpdate {
		persistedBlocklist := StoredListConfiguration{
			UpdateTimestamp: int(time.Now().Unix()),
			Blocklists:      blocklists,
			BlockedNames:    blockageMap,
		}
		persistedBlocklist.Persist(persistedBlocklistPath)
	}
}
