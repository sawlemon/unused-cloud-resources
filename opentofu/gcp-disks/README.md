# GCP OpenToFu module

## Authenticating

```hcl
gcloud auth application-default login
```

Change the `providers.tf` file to point to the ADC file path obtained from the above. In case this is running in a production environment on AWS look into [On-premises or another cloud provider](https://cloud.google.com/docs/authentication/provide-credentials-adc#on-prem)
