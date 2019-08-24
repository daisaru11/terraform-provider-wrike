module github.com/daisaru11/terraform-provider-wrike

go 1.12

replace github.com/daisaru11/wrike-go => ../wrike-go

require (
	github.com/daisaru11/wrike-go v0.0.0-00010101000000-000000000000
	github.com/hashicorp/terraform v0.12.7
)
