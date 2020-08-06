package k8s

import v1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"

func GetPostgresBackupCommand(bamboo *v1alpha1.Bamboo, version string) (postgresBackupCommand string) {

	postgresBackupCommand = "pg_dump -U " +
		bamboo.Spec.Datasource.Username +
		" -d" +
		bamboo.Spec.Datasource.Database +
		" > /var/lib/postgresql/data/$(date +\"time_%H-%M_date_%d-%m-%y\")_version_" + version + ".sql"
	return postgresBackupCommand

}
