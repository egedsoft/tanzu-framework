# airgapped-network

## Summary

Enable Tanzu CLI users to copy TKG images from one registry to another that does not have access to the internet.
This is done by following the next steps
1.Copy all associated TKG images to tars on a local disk using this plugin
2.Copy these tar files to the air-gapped network (using a USB drive or other mechanism)
3.Upload all the tar files to the air-gapped registry using this plugin

## Usage

### publish-image-to-tar

#### help output

```shell
$ tanzu airgapped-network publish-image-to-tar --help
Save image from public repository to tar files

Usage:
  tanzu airgapped-network publish-image-to-tar [flags]

Examples:
    # copy image from projects.registry.vmware.com/tkg to /tmp folder
    tanzu airgapped-network publish-image-to-tar --tkgImageRepository projects.registry.vmware.com/tkg --tkgVersion v1.6.0  --customImageRepo \<private repo\>


Flags:
      --tkgImageRepository          public repository path
      --tkgVersion                  tkg Version, which needs to be downloade
      --customImageRepo             custom repository path
  -h, --help                                help for image pull
```

### publish-image-from-tar

#### help output

```shell
$ tanzu airgapped-network publish-image-to-tar --help
Push images from tar files to private repository

Usage:
  tanzu airgapped-network publish-image-from-tar [flags]

Examples:
    # push images from tar file to private repository
    tanzu airgapped-network publish-image-from-tar --tkgTarFilePath /tmp   --customRepoCertificate ca.cer

Flags:
      --tkgTarFilePath  Absolute path of tar files
  -h, --help                                help for image push
```
