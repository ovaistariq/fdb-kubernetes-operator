/*
 * foundationdbbrestore_types_test.go
 *
 * This source file is part of the FoundationDB open source project
 *
 * Copyright 2020-2021 Apple Inc. and the FoundationDB project authors
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

package v1beta1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[api] FoundationDBRestore", func() {
	When("getting the backup URL", func() {
		DescribeTable("should generate the correct backup URL",
			func(restore FoundationDBRestore, expected string) {
				Expect(restore.BackupURL()).To(Equal(expected))
			},
			Entry("A restore with the backup url set",
				FoundationDBRestore{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mybackup",
					},
					Spec: FoundationDBRestoreSpec{
						BackupURL: "blobstore://test@test/mybackup?bucket=fdb-backups",
					},
				},
				"blobstore://test@test/mybackup?bucket=fdb-backups"),
			Entry("A restore with a blobstore config",
				FoundationDBRestore{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mybackup",
					},
					Spec: FoundationDBRestoreSpec{
						BlobStoreConfiguration: &BlobStoreConfiguration{
							AccountName: "account@account",
						},
					},
				},
				"blobstore://account@account/mybackup?bucket=fdb-backups"),
			Entry("A restore with a blobstore config with backup name",
				FoundationDBRestore{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mybackup",
					},
					Spec: FoundationDBRestoreSpec{
						BlobStoreConfiguration: &BlobStoreConfiguration{
							AccountName: "account@account",
							BackupName:  "test",
						},
					},
				},
				"blobstore://account@account/test?bucket=fdb-backups"),
			Entry("A restore with a blobstore config with a bucket name",
				FoundationDBRestore{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mybackup",
					},
					Spec: FoundationDBRestoreSpec{
						BlobStoreConfiguration: &BlobStoreConfiguration{
							AccountName: "account@account",
							Bucket:      "my-bucket",
						},
					},
				},
				"blobstore://account@account/mybackup?bucket=my-bucket"),
			Entry("A restore with a blobstore config with a bucket and backup name",
				FoundationDBRestore{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mybackup",
					},
					Spec: FoundationDBRestoreSpec{
						BlobStoreConfiguration: &BlobStoreConfiguration{
							AccountName: "account@account",
							BackupName:  "test",
							Bucket:      "my-bucket",
						},
					},
				},
				"blobstore://account@account/test?bucket=my-bucket"),
			Entry("A restore with a blobstore config with HTTP parameters and backup and bucket name",
				FoundationDBRestore{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mybackup",
					},
					Spec: FoundationDBRestoreSpec{
						BlobStoreConfiguration: &BlobStoreConfiguration{
							AccountName: "account@account",
							BackupName:  "test",
							Bucket:      "my-bucket",
							URLParameters: []URLParamater{
								"secure_connection=0",
							},
						},
					},
				},
				"blobstore://account@account/test?bucket=my-bucket&secure_connection=0"),
			Entry("A restore with a blobstore config with HTTP parameters",
				FoundationDBRestore{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mybackup",
					},
					Spec: FoundationDBRestoreSpec{
						BlobStoreConfiguration: &BlobStoreConfiguration{
							AccountName: "account@account",
							URLParameters: []URLParamater{
								"secure_connection=0",
							},
						},
					},
				},
				"blobstore://account@account/mybackup?bucket=fdb-backups&secure_connection=0"),
		)
	})
})
