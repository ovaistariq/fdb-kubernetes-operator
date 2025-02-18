/*
 * metrics.go
 *
 * This source file is part of the FoundationDB open source project
 *
 * Copyright 2021 Apple Inc. and the FoundationDB project authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	fdbtypes "github.com/FoundationDB/fdb-kubernetes-operator/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("metrics", func() {
	var cluster *fdbtypes.FoundationDBCluster

	BeforeEach(func() {
		cluster = &fdbtypes.FoundationDBCluster{
			Status: fdbtypes.FoundationDBClusterStatus{
				ProcessGroups: []*fdbtypes.ProcessGroupStatus{
					{
						ProcessClass: fdbtypes.ProcessClassStorage,
					},
					{
						ProcessClass: fdbtypes.ProcessClassLog,
						ProcessGroupConditions: []*fdbtypes.ProcessGroupCondition{
							fdbtypes.NewProcessGroupCondition(fdbtypes.MissingProcesses),
						},
					},
					{
						ProcessClass: fdbtypes.ProcessClassStorage,
						Remove:       true,
					},
					{
						ProcessClass: fdbtypes.ProcessClassStateless,
						Remove:       true,
						Excluded:     true,
					},
				},
			},
		}
	})

	Context("Collecting the processGroup metrics", func() {
		It("generate the process class metrics", func() {
			stats, removals, exclusions := getProcessGroupMetrics(cluster)
			Expect(len(stats)).To(BeNumerically("==", 3))
			Expect(len(stats[fdbtypes.ProcessClassStorage])).To(BeNumerically("==", len(fdbtypes.AllProcessGroupConditionTypes())))
			Expect(len(stats[fdbtypes.ProcessClassStorage])).To(BeNumerically("==", len(fdbtypes.AllProcessGroupConditionTypes())))
			Expect(stats[fdbtypes.ProcessClassStorage][fdbtypes.ReadyCondition]).To(BeNumerically("==", 2))
			Expect(stats[fdbtypes.ProcessClassLog][fdbtypes.ReadyCondition]).To(BeNumerically("==", 0))
			Expect(stats[fdbtypes.ProcessClassLog][fdbtypes.MissingProcesses]).To(BeNumerically("==", 1))
			Expect(stats[fdbtypes.ProcessClassStateless][fdbtypes.ReadyCondition]).To(BeNumerically("==", 1))
			Expect(removals[fdbtypes.ProcessClassStorage]).To(BeNumerically("==", 1))
			Expect(exclusions[fdbtypes.ProcessClassStorage]).To(BeNumerically("==", 0))
			Expect(removals[fdbtypes.ProcessClassStateless]).To(BeNumerically("==", 1))
			Expect(exclusions[fdbtypes.ProcessClassStateless]).To(BeNumerically("==", 1))
		})
	})
})
