
<a name="EdgeX Services (found in edgex-go) Changelog"></a>
## EdgeX Foundry Services
[Github repository](https://github.com/edgexfoundry/edgex-go)

## Change Logs for EdgeX Dependencies

- [go-mod-bootstrap](https://github.com/edgexfoundry/go-mod-bootstrap/blob/main/CHANGELOG.md)
- [go-mod-core-contracts](https://github.com/edgexfoundry/go-mod-core-contracts/blob/main/CHANGELOG.md)
- [go-mod-messaging](https://github.com/edgexfoundry/go-mod-messaging/blob/main/CHANGELOG.md)
- [go-mod-registry](https://github.com/edgexfoundry/go-mod-registry/blob/main/CHANGELOG.md) 
- [go-mod-secrets](https://github.com/edgexfoundry/go-mod-secrets/blob/main/CHANGELOG.md)
- [go-mod-configuration](https://github.com/edgexfoundry/go-mod-configuration/blob/main/CHANGELOG.md) (indirect dependency)

## [v2.3.0] Levski - 2022-11-09 (Only compatible with the 2.x release)

### Features ✨

- Secure MQTT Bus Credentials ([#4142](https://github.com/edgexfoundry/edgex-go/issues/4142)) ([#21503d9d](https://github.com/edgexfoundry/edgex-go/commits/21503d9d))
- Add secure mqtt support for eKuiper ([#4155](https://github.com/edgexfoundry/edgex-go/issues/4155)) ([#b7de7b77](https://github.com/edgexfoundry/edgex-go/commits/b7de7b77))
- Enable common security secret service metrics ([#4184](https://github.com/edgexfoundry/edgex-go/issues/4184)) ([#8629e80a](https://github.com/edgexfoundry/edgex-go/commits/8629e80a))
- Add go-winio to attribution (new SPIFFE dependency) ([#4029](https://github.com/edgexfoundry/edgex-go/issues/4029)) ([#ba7dda9b](https://github.com/edgexfoundry/edgex-go/commits/ba7dda9b))
- use different consul service policy rule for service role ([#4104](https://github.com/edgexfoundry/edgex-go/issues/4104)) ([#85bfaaf4](https://github.com/edgexfoundry/edgex-go/commits/85bfaaf4))
- Implement Device System Events ([#4101](https://github.com/edgexfoundry/edgex-go/issues/4101)) ([#643a1ec0](https://github.com/edgexfoundry/edgex-go/commits/643a1ec0))
- Add management acl token and save token info to file ([#4126](https://github.com/edgexfoundry/edgex-go/issues/4126)) ([#96e03a37](https://github.com/edgexfoundry/edgex-go/commits/96e03a37))
- Add option to build Core/Support services with NATS Capability ([#4137](https://github.com/edgexfoundry/edgex-go/issues/4137)) ([#604c3c95](https://github.com/edgexfoundry/edgex-go/commits/604c3c95))
- Add NATS Subject for autoprovisioning streams ([#4165](https://github.com/edgexfoundry/edgex-go/issues/4165)) ([#cf21bfd5](https://github.com/edgexfoundry/edgex-go/commits/cf21bfd5))
- Add security Consul metrics ([#4203](https://github.com/edgexfoundry/edgex-go/issues/4203)) ([#1656a83b](https://github.com/edgexfoundry/edgex-go/commits/1656a83b))
- **command:** 3rd party command query via messaging ([#4135](https://github.com/edgexfoundry/edgex-go/issues/4135)) ([#a0e04e55](https://github.com/edgexfoundry/edgex-go/commits/a0e04e55))
- **command:** Command request from external MQTT to internal MessageBus ([#4153](https://github.com/edgexfoundry/edgex-go/issues/4153)) ([#1c2713d9](https://github.com/edgexfoundry/edgex-go/commits/1c2713d9))
- **command:** Update core-command to support message bus access ([#4129](https://github.com/edgexfoundry/edgex-go/issues/4129)) ([#fd31c876](https://github.com/edgexfoundry/edgex-go/commits/fd31c876))
- **command:** Publish device service response to external MQTT ([#4166](https://github.com/edgexfoundry/edgex-go/issues/4166)) ([#1845739d](https://github.com/edgexfoundry/edgex-go/commits/1845739d))
- **command:** Subscribe command request via internal messaging ([#6f71b1dc](https://github.com/edgexfoundry/edgex-go/commits/6f71b1dc))
- **command:** Command query via internal messaging ([#4182](https://github.com/edgexfoundry/edgex-go/issues/4182)) ([#07636ef9](https://github.com/edgexfoundry/edgex-go/commits/07636ef9))
- **metadata:** Add UoM validation to POST deviceresource API ([#4140](https://github.com/edgexfoundry/edgex-go/issues/4140)) ([#f87a0e7a](https://github.com/edgexfoundry/edgex-go/commits/f87a0e7a))
- **metadata:** Implement GET Units of Measure API ([#4127](https://github.com/edgexfoundry/edgex-go/issues/4127)) ([#de42c75f](https://github.com/edgexfoundry/edgex-go/commits/de42c75f))
- **metadata:** Implement Unit of Measure (UoM) ADR ([#4119](https://github.com/edgexfoundry/edgex-go/issues/4119)) ([#03487ec6](https://github.com/edgexfoundry/edgex-go/commits/03487ec6))
- **snap:** Remove device-virtual ([#4041](https://github.com/edgexfoundry/edgex-go/issues/4041)) ([#fed831e0](https://github.com/edgexfoundry/edgex-go/commits/fed831e0))
- **snap:** Add secretstore token for device camera services ([#4034](https://github.com/edgexfoundry/edgex-go/issues/4034)) ([#6f84e707](https://github.com/edgexfoundry/edgex-go/commits/6f84e707))

### Bug Fixes 🐛

- Delayed start services for support- services fail ([#4159](https://github.com/edgexfoundry/edgex-go/issues/4159)) ([#726ff502](https://github.com/edgexfoundry/edgex-go/commits/726ff502))
- Fixed security-bootstrapper Docker volume init semantics ([#4085](https://github.com/edgexfoundry/edgex-go/issues/4085)) ([#ad21f989](https://github.com/edgexfoundry/edgex-go/commits/ad21f989))
- Fix security-secretstore-setup volume init semantics ([#4092](https://github.com/edgexfoundry/edgex-go/issues/4092)) ([#66f7195a](https://github.com/edgexfoundry/edgex-go/commits/66f7195a))
- Fixed Missing error check ([#4100](https://github.com/edgexfoundry/edgex-go/issues/4100)) ([#062af8d3](https://github.com/edgexfoundry/edgex-go/commits/062af8d3))
- Always run "kong migrations up" ([#4172](https://github.com/edgexfoundry/edgex-go/issues/4172)) ([#f2edda91](https://github.com/edgexfoundry/edgex-go/commits/f2edda91))
- Security services' bootstrap handler return true/false properly ([#4107](https://github.com/edgexfoundry/edgex-go/issues/4107)) ([#48044024](https://github.com/edgexfoundry/edgex-go/commits/48044024))
- Add consul policy list check ([#4164](https://github.com/edgexfoundry/edgex-go/issues/4164)) ([#a3948579](https://github.com/edgexfoundry/edgex-go/commits/a3948579))
- Add missing edgex/security config stem resolution in security-consul-bootstrapper ([#4110](https://github.com/edgexfoundry/edgex-go/issues/4110)) ([#a9913775](https://github.com/edgexfoundry/edgex-go/commits/a9913775))
- Addressed permission error in redis v7 [#4027](https://github.com/edgexfoundry/edgex-go/issues/4027) ([#4118](https://github.com/edgexfoundry/edgex-go/issues/4118)) ([#153a8924](https://github.com/edgexfoundry/edgex-go/commits/153a8924))
- Change redis username from redis5 to default ([#4031](https://github.com/edgexfoundry/edgex-go/issues/4031)) ([#4c165043](https://github.com/edgexfoundry/edgex-go/commits/4c165043))
- **data:** Fix deleting numerous events HTTP request timeout by goroutines ([#4063](https://github.com/edgexfoundry/edgex-go/issues/4063)) ([#e2bcede5](https://github.com/edgexfoundry/edgex-go/commits/e2bcede5))
- **scheduler:** Check interval when adding intervalAction ([#895df5c4](https://github.com/edgexfoundry/edgex-go/commits/895df5c4))
- **scheduler:** Remove unused field in scheduler config ([#81b3411d](https://github.com/edgexfoundry/edgex-go/commits/81b3411d))
- **snap:** Bind redis server to loopback interface ([#4154](https://github.com/edgexfoundry/edgex-go/issues/4154)) ([#0ef95e30](https://github.com/edgexfoundry/edgex-go/commits/0ef95e30))

### Code Refactoring ♻

- Move all the db related check to the infrastructure layer ([#b0f671d9](https://github.com/edgexfoundry/edgex-go/commits/b0f671d9))
- Move consul access and role interface ([#4193](https://github.com/edgexfoundry/edgex-go/issues/4193)) ([#5b680ce2](https://github.com/edgexfoundry/edgex-go/commits/5b680ce2))
- **command:** refactor messaging handler ([#4204](https://github.com/edgexfoundry/edgex-go/issues/4204)) ([#23251983](https://github.com/edgexfoundry/edgex-go/commits/23251983))
- **data:** Use deepCopy of messageBusInfo to avoid external adds ([#4038](https://github.com/edgexfoundry/edgex-go/issues/4038)) ([#9735311b](https://github.com/edgexfoundry/edgex-go/commits/9735311b))
- **snap:** Option handling and helper-go build ([#822255de](https://github.com/edgexfoundry/edgex-go/commits/822255de))
- **snap:** Merge install and configure packages ([#d8cdf693](https://github.com/edgexfoundry/edgex-go/commits/d8cdf693))
- **snap:** Environment variable usage and const definitions ([#4207](https://github.com/edgexfoundry/edgex-go/issues/4207)) ([#87ac77fe](https://github.com/edgexfoundry/edgex-go/commits/87ac77fe))
- **snap:** Use snapctl and log packages ([#4187](https://github.com/edgexfoundry/edgex-go/issues/4187)) ([#d47f91d7](https://github.com/edgexfoundry/edgex-go/commits/d47f91d7))

### Documentation 📖

- Publish swager to 2.3.0 ([#4023](https://github.com/edgexfoundry/edgex-go/issues/4023)) ([#f46214d8](https://github.com/edgexfoundry/edgex-go/commits/f46214d8))

### Build 👷

- Dockerfile fix ([#4217](https://github.com/edgexfoundry/edgex-go/issues/4217)) ([#8ec17376](https://github.com/edgexfoundry/edgex-go/commits/8ec17376))
- Upgrade CodeQL modules, enable additional checkers ([#4080](https://github.com/edgexfoundry/edgex-go/issues/4080)) ([#c49e1d00](https://github.com/edgexfoundry/edgex-go/commits/c49e1d00))
- Upgrade to go 1.18 and latest golangci-lint ([#4068](https://github.com/edgexfoundry/edgex-go/issues/4068)) ([#fb67e349](https://github.com/edgexfoundry/edgex-go/commits/fb67e349))
- Add convenient alias for each service's build and docker target ([#4123](https://github.com/edgexfoundry/edgex-go/issues/4123)) ([#23918450](https://github.com/edgexfoundry/edgex-go/commits/23918450))
- Optimize test-attribution-txt.sh to use go.mod, not vendor ([#4059](https://github.com/edgexfoundry/edgex-go/issues/4059)) ([#25cc1d46](https://github.com/edgexfoundry/edgex-go/commits/25cc1d46))
- Switch from docker to alpine base image for sys-mgmt-agent ([#6a305a0d](https://github.com/edgexfoundry/edgex-go/commits/6a305a0d))
- **snap:** Kong migrations up before Kong startup ([#4223](https://github.com/edgexfoundry/edgex-go/issues/4223)) ([#c8dc4aaa](https://github.com/edgexfoundry/edgex-go/commits/c8dc4aaa))
- **snap:** Exclude redundant eKuiper dependencies ([#4035](https://github.com/edgexfoundry/edgex-go/issues/4035)) ([#a09756d4](https://github.com/edgexfoundry/edgex-go/commits/a09756d4))
- **snap:** Exclude redundant eKuiper dependencies ([#8ccbfcee](https://github.com/edgexfoundry/edgex-go/commits/8ccbfcee))
- **snap:** Upgrade kong, redis, consul, vault ([#4210](https://github.com/edgexfoundry/edgex-go/issues/4210)) ([#bec6ac51](https://github.com/edgexfoundry/edgex-go/commits/bec6ac51))

### Other changes

- Update instructions to get latest linter ([#4067](https://github.com/edgexfoundry/edgex-go/issues/4067)) ([#33a3e4f3](https://github.com/edgexfoundry/edgex-go/commits/33a3e4f3))

### Continuous Integration 🔄

- Add blubracket ignore file ([#4167](https://github.com/edgexfoundry/edgex-go/issues/4167)) ([#a8eb28e7](https://github.com/edgexfoundry/edgex-go/commits/a8eb28e7))
- Add local cache to optimize local docker builds ([#eebbc51e](https://github.com/edgexfoundry/edgex-go/commits/eebbc51e))
- Update spire Dockerfile's to allow conditional compilation ([#184e7bf7](https://github.com/edgexfoundry/edgex-go/commits/184e7bf7))

## [v2.2.0] Kamakura - 2022-5-11 (Only compatible with the 2.x release)

### Features ✨
- **all:** Make support services include spiffe in buildtime from Makefile ([#3969](https://github.com/edgexfoundry/edgex-go/issues/3969)) ([#efde56b5](https://github.com/edgexfoundry/edgex-go/commits/efde56b5))
- **all:** Location of client service obtained from the registry ([#3879](https://github.com/edgexfoundry/edgex-go/issues/3879)) ([#2ec8c03a](https://github.com/edgexfoundry/edgex-go/commits/2ec8c03a))
- **all:** Implement service metrics for Events and Readings persisted ([#3922](https://github.com/edgexfoundry/edgex-go/issues/3922)) ([#e06225a2](https://github.com/edgexfoundry/edgex-go/commits/e06225a2))
- **all:** Create Common DTOs with ServiceName ([#3870](https://github.com/edgexfoundry/edgex-go/issues/3870)) ([#d6b89ac6](https://github.com/edgexfoundry/edgex-go/commits/d6b89ac6))
- **data:** Make MaxEventSize a service configuration setting ([#3891](https://github.com/edgexfoundry/edgex-go/issues/3891)) ([#de3e46c9](https://github.com/edgexfoundry/edgex-go/commits/de3e46c9))
- **metadata:** Implement Writable.ProfileChange configurations ([#a20eafb3](https://github.com/edgexfoundry/edgex-go/commits/a20eafb3))
- **metadata:** Implement PATCH deviceprofile/resource API ([#a40e21f6](https://github.com/edgexfoundry/edgex-go/commits/a40e21f6))
- **metadata:** Implement DELETE device command API ([#ac0e4352](https://github.com/edgexfoundry/edgex-go/commits/ac0e4352))
- **metadata:** Implement DELETE device resource API ([#691b12bf](https://github.com/edgexfoundry/edgex-go/commits/691b12bf))
- **metadata:** Implement PATCH /deviceprofile/deviceCommand API ([#0fa976f9](https://github.com/edgexfoundry/edgex-go/commits/0fa976f9))
- **metadata:** Invoke validation before adding/updating device ([#3900](https://github.com/edgexfoundry/edgex-go/issues/3900)) ([#b3afc0ae](https://github.com/edgexfoundry/edgex-go/commits/b3afc0ae))
- **metadata:** Implement Post device command API ([#dcb0ece3](https://github.com/edgexfoundry/edgex-go/commits/dcb0ece3))
- **metadata:** Implement PATCH device profile basic info API ([#243ad5ab](https://github.com/edgexfoundry/edgex-go/commits/243ad5ab))
- **metadata:** Implement POST device resource ([#3930](https://github.com/edgexfoundry/edgex-go/issues/3930)) ([#7587fe9d](https://github.com/edgexfoundry/edgex-go/commits/7587fe9d))
- **security:** Add secret store config for delayed start ([#3948](https://github.com/edgexfoundry/edgex-go/issues/3948)) ([#6b188fc4](https://github.com/edgexfoundry/edgex-go/commits/6b188fc4))
- **security:** Enable security-hardened go binaries for cgo flags ([#3893](https://github.com/edgexfoundry/edgex-go/issues/3893)) ([#7a573170](https://github.com/edgexfoundry/edgex-go/commits/7a573170))
- **security:** Implement spiffe token provider services ([#3897](https://github.com/edgexfoundry/edgex-go/issues/3897)) ([#81bad761](https://github.com/edgexfoundry/edgex-go/commits/81bad761))
- **security:** Update the pre-authorized service list for spire server config ([#3942](https://github.com/edgexfoundry/edgex-go/issues/3942)) ([#c716e684](https://github.com/edgexfoundry/edgex-go/commits/c716e684))
- **security:** Use conditional go build tags for delayed start and no_messagebus ([#3967](https://github.com/edgexfoundry/edgex-go/issues/3967)) ([#0af16247](https://github.com/edgexfoundry/edgex-go/commits/0af16247))
- **snap:** Add secretstore tokens for edgex-ekuiper ([#3888](https://github.com/edgexfoundry/edgex-go/issues/3888)) ([#d0bb8bce](https://github.com/edgexfoundry/edgex-go/commits/d0bb8bce))
- **snap:** Add additional tokens for app-service-configurable profiles ([#3825](https://github.com/edgexfoundry/edgex-go/issues/3825)) ([#23881e65](https://github.com/edgexfoundry/edgex-go/commits/23881e65))
- **snap:** Add support for environment variable injection ([#3986](https://github.com/edgexfoundry/edgex-go/issues/3986)) ([#780750f1](https://github.com/edgexfoundry/edgex-go/commits/780750f1))

### Bug Fixes 🐛
- **all:** Correct jwt error when reading private key ([#3843](https://github.com/edgexfoundry/edgex-go/issues/3843)) ([#1876cd19](https://github.com/edgexfoundry/edgex-go/commits/1876cd19))
- **all:** return 416 code when count is zero and offset > count ([#2c16b7b0](https://github.com/edgexfoundry/edgex-go/commits/2c16b7b0))
- **core-command:** Restore creation of DeviceServiceCommandClient in handler ([#89cc48a7](https://github.com/edgexfoundry/edgex-go/commits/89cc48a7))
- **metadata:** Fix the typo of notification message ([#5146f317](https://github.com/edgexfoundry/edgex-go/commits/5146f317))
- **metadata:** Not trigger DS callback if only lastConnected is updated ([#3976](https://github.com/edgexfoundry/edgex-go/issues/3976)) ([#7c172932](https://github.com/edgexfoundry/edgex-go/commits/7c172932))
- **metadata:** Add 404 case for not-yet-upgraded device-service ([#79655804](https://github.com/edgexfoundry/edgex-go/commits/79655804))
- **metadata:** Ignore validation result if device service unavailable ([#b9439831](https://github.com/edgexfoundry/edgex-go/commits/b9439831))
- **security:** Security-proxy-setup will not terminate on success ([#3917](https://github.com/edgexfoundry/edgex-go/issues/3917)) ([#d0a5aad2](https://github.com/edgexfoundry/edgex-go/commits/d0a5aad2))
- **security:** Fix for spiffe-config crashloop ([#3947](https://github.com/edgexfoundry/edgex-go/issues/3947)) ([#1a2a9fb2](https://github.com/edgexfoundry/edgex-go/commits/1a2a9fb2))
- **sma:** Update Docker base image in sys mgmt agent Dockerfile ([#3847](https://github.com/edgexfoundry/edgex-go/issues/3847)) ([#9e663fa0](https://github.com/edgexfoundry/edgex-go/commits/9e663fa0))
- **snap:** Apply proxy's runtime config options after startup ([#3856](https://github.com/edgexfoundry/edgex-go/issues/3856)) ([#3825f82a](https://github.com/edgexfoundry/edgex-go/commits/3825f82a))
- **snap:** Deploy secrets token into separate mount points ([#3826](https://github.com/edgexfoundry/edgex-go/issues/3826)) ([#b0e5e643](https://github.com/edgexfoundry/edgex-go/commits/b0e5e643))

### Code Refactoring ♻
- Rename Encode func to better name ([#3822](https://github.com/edgexfoundry/edgex-go/issues/3822)) ([#b2bbeb2b](https://github.com/edgexfoundry/edgex-go/commits/b2bbeb2b))
- **all:** Use formatted alternative log function instead of fmt.Sprintf ([#46183a66](https://github.com/edgexfoundry/edgex-go/commits/46183a66))
- **data:** Move subscriber.go from application to controller ([#3903](https://github.com/edgexfoundry/edgex-go/issues/3903)) ([#55644671](https://github.com/edgexfoundry/edgex-go/commits/55644671))
- **metadata:** Move DeleteDeviceResourceByName and DeleteDeviceCommandByName to correct files ([#80f156bd](https://github.com/edgexfoundry/edgex-go/commits/80f156bd))

### Documentation 📖
- Add 503 response code to core-command API document ([#3839](https://github.com/edgexfoundry/edgex-go/issues/3839)) ([#f0bf1273](https://github.com/edgexfoundry/edgex-go/commits/f0bf1273))
- Update core-metadata API doc PATCH/PUT examples ([#0cef4b84](https://github.com/edgexfoundry/edgex-go/commits/0cef4b84))
- Update ([#c6553be1](https://github.com/edgexfoundry/edgex-go/commits/c6553be1))
- Update issue templates to indicate required sections ([#95e17dcf](https://github.com/edgexfoundry/edgex-go/commits/95e17dcf))
- Add path field to support-scheduler intervalAction API example ([#8b40f194](https://github.com/edgexfoundry/edgex-go/commits/8b40f194))
- **all:** Upgrade the API Docs version to 2.2.0 ([#3857](https://github.com/edgexfoundry/edgex-go/issues/3857)) ([#1e28c249](https://github.com/edgexfoundry/edgex-go/commits/1e28c249))
- **metadata:** Update metadata swagger API docs for device profile changes ([#a6fe8a26](https://github.com/edgexfoundry/edgex-go/commits/a6fe8a26))
- **metadata:** Fix Swagger metadata PATCH deviceprofile/resource response ([#a6024f17](https://github.com/edgexfoundry/edgex-go/commits/a6024f17))
- **metadata:** Update metadata Swagger PATCH API responses ([#380509e6](https://github.com/edgexfoundry/edgex-go/commits/380509e6))
- **snap:** Move usage instructions to docs ([#4006](https://github.com/edgexfoundry/edgex-go/issues/4006)) ([#75ae4c63](https://github.com/edgexfoundry/edgex-go/commits/75ae4c63))
- **snap:** Update env config overrides description ([#35f1a85d](https://github.com/edgexfoundry/edgex-go/commits/35f1a85d))
- **snap:** Update env configuration overrides ([#50ed3c20](https://github.com/edgexfoundry/edgex-go/commits/50ed3c20))
- **snap:** Revert to `csv` in snap/README.md ([#023217a9](https://github.com/edgexfoundry/edgex-go/commits/023217a9))

### Build 👷
- Update to latest go-mod-messaging w/o ZMQ on windows ([#3833](https://github.com/edgexfoundry/edgex-go/issues/3833)) ([#240e6fd7](https://github.com/edgexfoundry/edgex-go/commits/240e6fd7))
    ```
    BREAKING CHANGE:
    ZeroMQ no longer supported on native Windows for EdgeX
    MessageBus
    ```
- Update to latest go-mod-messaging with ZMQ fix ([#3876](https://github.com/edgexfoundry/edgex-go/issues/3876)) ([#6e64952c](https://github.com/edgexfoundry/edgex-go/commits/6e64952c))
- Updated formating from gofmt 1.17 ([#3861](https://github.com/edgexfoundry/edgex-go/issues/3861)) ([#76a2231e](https://github.com/edgexfoundry/edgex-go/commits/76a2231e))
- **snap:** Upgrade kuiper to 1.4.2 ([#3867](https://github.com/edgexfoundry/edgex-go/issues/3867)) ([#85bf44e2](https://github.com/edgexfoundry/edgex-go/commits/85bf44e2))
- **snap:** Upgrade to device-virtual v2.2.0-dev.7 and app-service-configurable latest/beta ([#3914](https://github.com/edgexfoundry/edgex-go/issues/3914)) ([#7794663b](https://github.com/edgexfoundry/edgex-go/commits/7794663b))
- **snap:** Source metadata from central repo ([#3959](https://github.com/edgexfoundry/edgex-go/issues/3959)) ([#ad3b9b7a](https://github.com/edgexfoundry/edgex-go/commits/ad3b9b7a))
- **snap:** Update edgexfoundry snap base to core20 ([#3848](https://github.com/edgexfoundry/edgex-go/issues/3848)) ([#2940d6a5](https://github.com/edgexfoundry/edgex-go/commits/2940d6a5))

### Continuous Integration 🔄
- Go 1.17 related changes ([#3864](https://github.com/edgexfoundry/edgex-go/issues/3864)) ([#da06b860](https://github.com/edgexfoundry/edgex-go/commits/da06b860))

## [v2.1.0] Jakarta - 2021-11-17 (Only compatible with the 2.x release)

### Features ✨
- **all:** Enable CORS headers for all services ([#3758](https://github.com/edgexfoundry/edgex-go/issues/3758)) ([#4fce4fd2](https://github.com/edgexfoundry/edgex-go/commits/4fce4fd2))
- **all:** Update multi instance response to have totalCount field ([#ea5a8f40](https://github.com/edgexfoundry/edgex-go/commits/ea5a8f40))
- **command:** Support object value type in Set Command ([#eaa9784a](https://github.com/edgexfoundry/edgex-go/commits/eaa9784a))
- **command:** Update Command V2 API to include totalCount field for MultiDeviceCoreCommandsResponse ([#4ad05991](https://github.com/edgexfoundry/edgex-go/commits/4ad05991))
- **data:** Implement V2 API to query readings by name and time range ([#3577](https://github.com/edgexfoundry/edgex-go/issues/3577)) ([#8a6c1010](https://github.com/edgexfoundry/edgex-go/commits/8a6c1010))
- **data:** new API to search Readings by multiple resource names ([#3766](https://github.com/edgexfoundry/edgex-go/issues/3766)) ([#e2d5d6cc](https://github.com/edgexfoundry/edgex-go/commits/e2d5d6cc))
- **data:** Enhance the performance of events deletion ([#3722](https://github.com/edgexfoundry/edgex-go/issues/3722)) ([#2de07aa5](https://github.com/edgexfoundry/edgex-go/commits/2de07aa5))
- **data:** Support Object value type in Reading ([#94769bcc](https://github.com/edgexfoundry/edgex-go/commits/94769bcc))
- **data:** refactor application-layer multi-events func to reduce code duplication ([#753b88f4](https://github.com/edgexfoundry/edgex-go/commits/753b88f4))
- **data:** Update MultiReadingsResponse to have totalCount field ([#07c09b9a](https://github.com/edgexfoundry/edgex-go/commits/07c09b9a))
- **data:** Update MultiEventsResponse to have totalCount field ([#d627eae0](https://github.com/edgexfoundry/edgex-go/commits/d627eae0))
- **data:** implement new GET Readings API ([#1ef40f49](https://github.com/edgexfoundry/edgex-go/commits/1ef40f49))
- **metadata:** Send notification after updating device entity ([#3623](https://github.com/edgexfoundry/edgex-go/issues/3623)) ([#166d7917](https://github.com/edgexfoundry/edgex-go/commits/166d7917))
- **metadata:** Update Metadata V2 API to include totalCount field for multi-instance response ([#377c2adc](https://github.com/edgexfoundry/edgex-go/commits/377c2adc))
- **notifications:** Update Notififcation V2 API to include totalCount field ([#b1707c08](https://github.com/edgexfoundry/edgex-go/commits/b1707c08))
- **notifications:** add new API to Get Transmissions by Notification id ([#3759](https://github.com/edgexfoundry/edgex-go/issues/3759)) ([#4de7b29e](https://github.com/edgexfoundry/edgex-go/commits/4de7b29e))
- **scheduler:** Validate Interval and IntervalAction before loading from config ([#3646](https://github.com/edgexfoundry/edgex-go/issues/3646)) ([#c934d262](https://github.com/edgexfoundry/edgex-go/commits/c934d262))
- **scheduler:** Update Scheduler V2 API to include totalCount field ([#2b972191](https://github.com/edgexfoundry/edgex-go/commits/2b972191))
- **security:** Add injection of Secure MessageBus creds for eKuiper connections ([#3778](https://github.com/edgexfoundry/edgex-go/issues/3778)) ([#fb769a00](https://github.com/edgexfoundry/edgex-go/commits/fb769a00))
- **security:** Add Secret File config setting ([#3788](https://github.com/edgexfoundry/edgex-go/issues/3788)) ([#adab5248](https://github.com/edgexfoundry/edgex-go/commits/adab5248))
- **security:** Enable modern cipher suite / TLSv1.3 only ([#3704](https://github.com/edgexfoundry/edgex-go/issues/3704)) ([#7380b5be](https://github.com/edgexfoundry/edgex-go/commits/7380b5be))
- **security:** Make Vault token TTL configurable ([#3675](https://github.com/edgexfoundry/edgex-go/issues/3675)) ([#19484f48](https://github.com/edgexfoundry/edgex-go/commits/19484f48))
- **snap:** add vault ttl config support ([#ef3901f9](https://github.com/edgexfoundry/edgex-go/commits/ef3901f9))
- **snap:** add additional devices to secret store lists in install hook ([#8ad81a0f](https://github.com/edgexfoundry/edgex-go/commits/8ad81a0f))

### Performance Improvements ⚡
- Change MaxResultCount setting to 1024 ([#8524b20a](https://github.com/edgexfoundry/edgex-go/commits/8524b20a))

### Bug Fixes 🐛
- **all:** http response cannot be completed ([#3662](https://github.com/edgexfoundry/edgex-go/issues/3662)) ([#0ba6ba5b](https://github.com/edgexfoundry/edgex-go/commits/0ba6ba5b))
- **command:** Using the Device Service response code for Get Command ([#9f422825](https://github.com/edgexfoundry/edgex-go/commits/9f422825))
- **command:** clean out database section from core command ([#0fae9ab3](https://github.com/edgexfoundry/edgex-go/commits/0fae9ab3))
- **command:** Fix core-command crashes error ([#86f6abfe](https://github.com/edgexfoundry/edgex-go/commits/86f6abfe))
- **data:** add codes to remove db index of reading:deviceName:resourceName when deleting readings ([#173b0957](https://github.com/edgexfoundry/edgex-go/commits/173b0957))
- **metadata:** Remove operating state from device service ([#dc27294b](https://github.com/edgexfoundry/edgex-go/commits/dc27294b))
- **metadata:** Disable device notification by default ([#3789](https://github.com/edgexfoundry/edgex-go/issues/3789)) ([#c5f5ac19](https://github.com/edgexfoundry/edgex-go/commits/c5f5ac19))
- **metadata:** device yaml marshal to Json  error ([#3683](https://github.com/edgexfoundry/edgex-go/issues/3683)) ([#e89d87e1](https://github.com/edgexfoundry/edgex-go/commits/e89d87e1))
- **metadata:** add labels as part of query criteria when finding ([#3781](https://github.com/edgexfoundry/edgex-go/issues/3781)) ([#11dac8c4](https://github.com/edgexfoundry/edgex-go/commits/11dac8c4))
- **security:** Move JWT auth method to individual routes ([#3657](https://github.com/edgexfoundry/edgex-go/issues/3657)) ([#d2a5f5fe](https://github.com/edgexfoundry/edgex-go/commits/d2a5f5fe))
- **security:** Replace abandoned JWT package ([#3729](https://github.com/edgexfoundry/edgex-go/issues/3729)) ([#32c3a59f](https://github.com/edgexfoundry/edgex-go/commits/32c3a59f))
- **security:** use localhost for kuiper config ([#8fa67b54](https://github.com/edgexfoundry/edgex-go/commits/8fa67b54))
- **security:** secrets-config user connect using TLS ([#3698](https://github.com/edgexfoundry/edgex-go/issues/3698)) ([#258ae4e0](https://github.com/edgexfoundry/edgex-go/commits/258ae4e0))
- **security:** remove unused curl executable from secretstore-setup Dockerfile - curl command executable is not used, so it is removed from the Docker file of service secretstore-setup ([#49239b82](https://github.com/edgexfoundry/edgex-go/commits/49239b82))
- **security:** Mismatched types int and int32 ([#3655](https://github.com/edgexfoundry/edgex-go/issues/3655)) ([#dbae55fc](https://github.com/edgexfoundry/edgex-go/commits/dbae55fc))
- **snap:** Update edgex-snap-hooks to v2.0.5 ([#6106f566](https://github.com/edgexfoundry/edgex-go/commits/6106f566))
- **snap:** Configure kuiper's REST service port ([#3770](https://github.com/edgexfoundry/edgex-go/issues/3770)) ([#a2b69b26](https://github.com/edgexfoundry/edgex-go/commits/a2b69b26))
- **snap:** Make secretstore connect hook idempotent ([#3815](https://github.com/edgexfoundry/edgex-go/issues/3815)) ([#387e8ab1](https://github.com/edgexfoundry/edgex-go/commits/387e8ab1))
- **snap:** Fix app-rules-engine ([#651aaa83](https://github.com/edgexfoundry/edgex-go/commits/651aaa83))
- **snap:** Add kuiper message-bus config ([#602d7f53](https://github.com/edgexfoundry/edgex-go/commits/602d7f53))
- **snap:** Remove config logic from install hook ([#af5ffab8](https://github.com/edgexfoundry/edgex-go/commits/af5ffab8))
- **snap:** Defer service start on install ([#0e26993c](https://github.com/edgexfoundry/edgex-go/commits/0e26993c))
- **snap:** Create empty redis.conf file ([#793f35af](https://github.com/edgexfoundry/edgex-go/commits/793f35af))
- **snap:** Support app/dev services w/secretstore=off ([#e36a01fe](https://github.com/edgexfoundry/edgex-go/commits/e36a01fe))
- **snap:** Support non-secure consul ([#9375f9f3](https://github.com/edgexfoundry/edgex-go/commits/9375f9f3))
- **snap:** Support disabling security ([#6f85a457](https://github.com/edgexfoundry/edgex-go/commits/6f85a457))

### Code Refactoring ♻
- **all:** Clean up TOML quotes ([#3666](https://github.com/edgexfoundry/edgex-go/issues/3666)) ([#729eb473](https://github.com/edgexfoundry/edgex-go/commits/729eb473))
- **all:** Refactor io.Reader for reusing ([#3627](https://github.com/edgexfoundry/edgex-go/issues/3627)) ([#7434bcad](https://github.com/edgexfoundry/edgex-go/commits/7434bcad))
- **all:** Remove unused Redis client variables ([#905a639d](https://github.com/edgexfoundry/edgex-go/commits/905a639d))
- **snap:** Convert store lists into slice in install hook ([#e09eec38](https://github.com/edgexfoundry/edgex-go/commits/e09eec38))

## [v2.0.0] Ireland - 2021-06-30  (Not Compatible with 1.x releases)

## General
- **v2:** Implemented Core Data V2 APIs as defined in [SwaggerHub](https://app.swaggerhub.com/apis/EdgeXFoundry1/core-data/2.x)
- **v2:** Implemented Core Command V2 APIs as defined in [SwaggerHub](https://app.swaggerhub.com/apis/EdgeXFoundry1/core-command/2.x)
- **v2:** Implemented Core Metadata V2 APIs as defined in [SwaggerHub](https://app.swaggerhub.com/apis/EdgeXFoundry1/core-metadata/2.x)
- **v2:** Implemented Support Scheduler V2 APIs as defined in [SwaggerHub](https://app.swaggerhub.com/apis/EdgeXFoundry1/support-scheduler/2.x)
- **v2:** Implemented Support Notifications V2 APIs as defined in [SwaggerHub](https://app.swaggerhub.com/apis/EdgeXFoundry1/support-notifications/2.x)
- **v2:** Implemented System Management Agent V2 APIs as defined in [SwaggerHub](https://app.swaggerhub.com/apis/EdgeXFoundry1/system-agent/2.x)
- **v2:** Change the default ports for EdgeX services to stay within [IANA Dynamic Ports](https://tools.ietf.org/id/draft-cotton-tsvwg-iana-ports-00.html#privateports)
- **v2:** Updated all Docker image names (removing docker- prefix and language suffixes of -go and -c)

### Features ✨
- **v2:** Remove --useradd and --userdel support from proxy-setup ([#2924](https://github.com/edgexfoundry/edgex-go/issues/2924)) ([#60451040](https://github.com/edgexfoundry/edgex-go/commits/60451040))
- **v2:** Processing query params of url in put method ([#3034](https://github.com/edgexfoundry/edgex-go/issues/3034)) ([#5c263209](https://github.com/edgexfoundry/edgex-go/commits/5c263209))
- **v2:** Configure Kuiper for secure message bus ([#3537](https://github.com/edgexfoundry/edgex-go/issues/3537)) ([#71bb76d4](https://github.com/edgexfoundry/edgex-go/commits/71bb76d4))
- **v2:** Use service keys for Route configuration keys ([#3247](https://github.com/edgexfoundry/edgex-go/issues/3247)) ([#c48b5c69](https://github.com/edgexfoundry/edgex-go/commits/c48b5c69))
- **v2:** Remove security services initialization for mongodb ([#2885](https://github.com/edgexfoundry/edgex-go/issues/2885)) ([#bd94ef45](https://github.com/edgexfoundry/edgex-go/commits/bd94ef45))
- **v2:** Enable the check of adminState for notifications and scheduler ([#33c15794](https://github.com/edgexfoundry/edgex-go/commits/33c15794))
- **v2:** Add missing middleware func to router ([#768023b2](https://github.com/edgexfoundry/edgex-go/commits/768023b2))
- **v2:** Remove deprecated Mongo code. ([#2956](https://github.com/edgexfoundry/edgex-go/issues/2956)) ([#dd265b0a](https://github.com/edgexfoundry/edgex-go/commits/dd265b0a))
- **v2:** Add RedisDB Password for v2 security mode ([#cbc1041f](https://github.com/edgexfoundry/edgex-go/commits/cbc1041f))
- **v2:** Remove MetadataCheck mechanism when adding Event ([#3069](https://github.com/edgexfoundry/edgex-go/issues/3069)) ([#f7cba1f5](https://github.com/edgexfoundry/edgex-go/commits/f7cba1f5))
- **v2:** Add secure MessageBus capability ([#3436](https://github.com/edgexfoundry/edgex-go/issues/3436)) ([#55d4d9f0](https://github.com/edgexfoundry/edgex-go/commits/55d4d9f0))
commits/9055af8f))
- **data:** Make Core Data publish events to <TopicPrefix>/<DeviceProfile>/<DeviceName> ([#3002](https://github.com/edgexfoundry/edgex-go/issues/3002)) ([#cd24e070](https://github.com/edgexfoundry/edgex-go/commits/cd24e070))
e47b23dc))
- **data:** Modify event validation error message ([#43e7fdfd](https://github.com/edgexfoundry/edgex-go/commits/43e7fdfd))
- **data:** Remove pushed field completely from V2 Event related implementation ([#f3d77c85](https://github.com/edgexfoundry/edgex-go/commits/f3d77c85))
- **data:** Add the missing event's sourceName at persistent layer ([#b7db4934](https://github.com/edgexfoundry/edgex-go/commits/b7db4934))
- **data:** Message topic should contain the event's deviceName ([#16398693](https://github.com/edgexfoundry/edgex-go/commits/16398693))
- **data:** Implement get Binary Reading from database ([#3303](https://github.com/edgexfoundry/edgex-go/issues/3303)) ([#d1fc5940](https://github.com/edgexfoundry/edgex-go/commits/d1fc5940))
- **data:** Remove created field from Event and Reading ([#3299](https://github.com/edgexfoundry/edgex-go/issues/3299)) ([#04121680](https://github.com/edgexfoundry/edgex-go/commits/04121680))
- **data:** Core Data remove V2 Pushed and Scrub APIs ([#33b5724a](https://github.com/edgexfoundry/edgex-go/commits/33b5724a))
- **notifications:** Check Subscription with empty categories,labels ([#45699a18](https://github.com/edgexfoundry/edgex-go/commits/45699a18))
- **notifications:** Add secret creation API ([#3510](https://github.com/edgexfoundry/edgex-go/issues/3510)) ([#20e30386](https://github.com/edgexfoundry/edgex-go/commits/20e30386))
- **notifications:** Implement Sending Service for Email Channel ([#3530](https://github.com/edgexfoundry/edgex-go/issues/3530)) ([#399b1e1f](https://github.com/edgexfoundry/edgex-go/commits/399b1e1f))
- **scheduler:** ServiceName change should invoke old service's callback ([#638c5eca](https://github.com/edgexfoundry/edgex-go/commits/638c5eca))
- **security:** Add new implementation for security bootstrapping/installation ([#2970](https://github.com/edgexfoundry/edgex-go/issues/2970)) ([#5dc76a6c](https://github.com/edgexfoundry/edgex-go/commits/5dc76a6c))
- **security:** Secure containers run as non-root ([#3003](https://github.com/edgexfoundry/edgex-go/issues/3003)) ([#310fcf06](https://github.com/edgexfoundry/edgex-go/commits/310fcf06))
- **security:** Implementation to set up Consul ACL ([#3215](https://github.com/edgexfoundry/edgex-go/issues/3215)) ([#8a562533](https://github.com/edgexfoundry/edgex-go/commits/8a562533))
- **security:** Create a Vault mgmt token for Consul Secrets API Operations ([#3192](https://github.com/edgexfoundry/edgex-go/issues/3192)) ([#257616ab](https://github.com/edgexfoundry/edgex-go/commits/257616ab))
- **security:** Implementation for setting up agent token ([#3251](https://github.com/edgexfoundry/edgex-go/issues/3251)) ([#7baeca4e](https://github.com/edgexfoundry/edgex-go/commits/7baeca4e))
- **security:** Add waitFor subcommand for security-bootstrapper ([#3101](https://github.com/edgexfoundry/edgex-go/issues/3101)) ([#f32f4191](https://github.com/edgexfoundry/edgex-go/commits/f32f4191))
- **security:** Implementation for generating consul tokens ([#3324](https://github.com/edgexfoundry/edgex-go/issues/3324)) ([#9479b0bd](https://github.com/edgexfoundry/edgex-go/commits/9479b0bd))
- **security:** Integrate EdgeX core servcies/app service with Consul tokens ([#3331](https://github.com/edgexfoundry/edgex-go/issues/3331)) ([#70f8294d](https://github.com/edgexfoundry/edgex-go/commits/70f8294d))
- **security:** Implement secrets-config proxy tls ([#2930](https://github.com/edgexfoundry/edgex-go/issues/2930)) ([#382321cd](https://github.com/edgexfoundry/edgex-go/commits/382321cd))
- **security:** Replace security-proxy-setup for adding users ([#2808](https://github.com/edgexfoundry/edgex-go/issues/2808)) ([#ff93af41](https://github.com/edgexfoundry/edgex-go/commits/ff93af41))
- **security:** Implement Consul token header in API Gateway ([#3391](https://github.com/edgexfoundry/edgex-go/issues/3391)) ([#58f175f3](https://github.com/edgexfoundry/edgex-go/commits/58f175f3))
- **security:** Secure Kong Admin API ([#3328](https://github.com/edgexfoundry/edgex-go/issues/3328)) ([#073d4024](https://github.com/edgexfoundry/edgex-go/commits/073d4024))
### Bug Fixes 🐛
- **security:** Enable Vault's Consul secrets engine ([#3179](https://github.com/edgexfoundry/edgex-go/issues/3179)) ([#13b869e2](https://github.com/edgexfoundry/edgex-go/commits/13b869e2))
- **all:** Invoke DS deletion Callback by name ([#b818cb7f](https://github.com/edgexfoundry/edgex-go/commits/b818cb7f))
- **all:** Added Content-TYpe from REST header to Context ([#c433a97c](https://github.com/edgexfoundry/edgex-go/commits/c433a97c))
- **metadata:** Check the provisionWatcher existence when delete DS ([#7014d8db](https://github.com/edgexfoundry/edgex-go/commits/7014d8db))
- **metadata:** Delete DS API should check the associated Device existence ([#3054](https://github.com/edgexfoundry/edgex-go/issues/3054)) ([#b641f4fe](https://github.com/edgexfoundry/edgex-go/commits/b641f4fe))
- **metadata:** Fix DS callback function panic error ([#3523](https://github.com/edgexfoundry/edgex-go/issues/3523)) ([#e6c05256](https://github.com/edgexfoundry/edgex-go/commits/e6c05256))
- **metadata:** Check the associated object existence when delete Profile ([#35d7beb0](https://github.com/edgexfoundry/edgex-go/commits/35d7beb0))
- **notifications:** Query with invalid offset range should return 416 ([#73839868](https://github.com/edgexfoundry/edgex-go/commits/73839868))
- **notifications:** Fix using the wrong key to update Subscription ([#fe00754a](https://github.com/edgexfoundry/edgex-go/commits/fe00754a))
- **notifications:** Return 200 when successful to delete subscription ([#cd63672c](https://github.com/edgexfoundry/edgex-go/commits/cd63672c))
- **scheduler:** PATCH API Check intervalName nil value ([#f5098ad0](https://github.com/edgexfoundry/edgex-go/commits/f5098ad0))
- **snap:** Cleanup snap hooks ([#9c984eaf](https://github.com/edgexfoundry/edgex-go/commits/9c984eaf))
- **snap:** Update device-virtual --confdir ([#ed9dddb0](https://github.com/edgexfoundry/edgex-go/commits/ed9dddb0))
- **snap:** Fix redis snapshots ([#187bb8da](https://github.com/edgexfoundry/edgex-go/commits/187bb8da))
- **snap:** Update pipe optimization patch ([#ce6ab2ee](https://github.com/edgexfoundry/edgex-go/commits/ce6ab2ee))
- **snap:** Update service command-line options ([#b880d5f7](https://github.com/edgexfoundry/edgex-go/commits/b880d5f7))
- **snap:** Update snap optimization patch ([#1010d7ab](https://github.com/edgexfoundry/edgex-go/commits/1010d7ab))
- **snap:** Remove kong TLS config overrides ([#2953](https://github.com/edgexfoundry/edgex-go/issues/2953)) ([#57027b83](https://github.com/edgexfoundry/edgex-go/commits/57027b83))
- **snap:** Secure kong admin ports ([#0985e833](https://github.com/edgexfoundry/edgex-go/commits/0985e833))
### Code Refactoring ♻
- **security:** Remove Writable from security services ([#3147](https://github.com/edgexfoundry/edgex-go/issues/3147)) ([#4701403b](https://github.com/edgexfoundry/edgex-go/commits/4701403b))
- **security:** Implementation for adding registry role on the fly ([#3291](https://github.com/edgexfoundry/edgex-go/issues/3291)) ([#18e95d4e](https://github.com/edgexfoundry/edgex-go/commits/18e95d4e))
- **security:** Fix Redis start issue from [#2863](https://github.com/edgexfoundry/edgex-go/issues/2863) ([#3115](https://github.com/edgexfoundry/edgex-go/issues/3115)) ([#cb6997bb](https://github.com/edgexfoundry/edgex-go/commits/cb6997bb))
- **security:** Eliminate security-secrets-setup module ([#2913](https://github.com/edgexfoundry/edgex-go/issues/2913)) ([#4ced080b](https://github.com/edgexfoundry/edgex-go/commits/4ced080b))
- **security:** Deprecate oauth2 auth method ([#3575](https://github.com/edgexfoundry/edgex-go/issues/3575)) ([#649de808](https://github.com/edgexfoundry/edgex-go/commits/649de808))
- **security:** Implementation for adding ACL policies and roles ([#3273](https://github.com/edgexfoundry/edgex-go/issues/3273)) ([#8b8c0450](https://github.com/edgexfoundry/edgex-go/commits/8b8c0450))
- **security:** Kong cert paths are now optional ([#2940](https://github.com/edgexfoundry/edgex-go/issues/2940)) ([#c80d9cda](https://github.com/edgexfoundry/edgex-go/commits/c80d9cda))
- **security:** Remove Vault dependency on Consul by using file backend ([#2886](https://github.com/edgexfoundry/edgex-go/issues/2886)) ([#f9701ca4](https://github.com/edgexfoundry/edgex-go/commits/f9701ca4))
- **v2:** Remove obsolete V1 code, swagger & scripts ([#3492](https://github.com/edgexfoundry/edgex-go/issues/3492)) ([#f99dd52e](https://github.com/edgexfoundry/edgex-go/commits/f99dd52e))
- **sma:** Use service key for route names and add missing sys-mgmt-agent ([#580caa8f](https://github.com/edgexfoundry/edgex-go/commits/580caa8f))
    ```
    BREAKING CHANGE:
    API Gateway route names have changed.
    ```
- **all:** Update Service configurations for changes to common Service configuration struct ([#e3cbfe1c](https://github.com/edgexfoundry/edgex-go/commits/e3cbfe1c))
    ```
    BREAKING CHANGE:
    Core/Support/SysMgmt Service configuration has changed.
    ```
- **all:** Assign/use new Port assignments ([#3485](https://github.com/edgexfoundry/edgex-go/issues/3485)) ([#1b47f7a0](https://github.com/edgexfoundry/edgex-go/commits/1b47f7a0))
    ```
    BREAKING CHANGE:
    Core/Support/SysMgmt default ports numbers have changed.
    ```
- **security:** Use new enhanced SecretProvider ([#2942](https://github.com/edgexfoundry/edgex-go/issues/2942)) ([#c8e01228](https://github.com/edgexfoundry/edgex-go/commits/c8e01228))
- **all:** Use latest bootstrap for logging client changes ([#2975](https://github.com/edgexfoundry/edgex-go/issues/2975)) ([#f96f7b91](https://github.com/edgexfoundry/edgex-go/commits/f96f7b91))
- **all:** Update for new service key names and overrides for hyphen to underscore ([#3462](https://github.com/edgexfoundry/edgex-go/issues/3462)) ([#e79253a2](https://github.com/edgexfoundry/edgex-go/commits/e79253a2))
    ```
    BREAKING CHANGE:
    Service key names used in configuration have changed.
    ```
- **scheduler:** Rename Frquency to Interval in AutoEvent and Scheduler ([#1ca8f34a](https://github.com/edgexfoundry/edgex-go/commits/1ca8f34a))
- **all:** Combine and improve http LoggingMiddleware ([#0f2753e7](https://github.com/edgexfoundry/edgex-go/commits/0f2753e7))
- **all:** Refactor controller error handling ([#3400](https://github.com/edgexfoundry/edgex-go/issues/3400)) ([#d6c94b2b](https://github.com/edgexfoundry/edgex-go/commits/d6c94b2b))
- **all:** Replace use of BurntSushi/toml with pelletier/go-toml ([#fa8052bc](https://github.com/edgexfoundry/edgex-go/commits/fa8052bc))
- **security:** Rework entry point scripts to run 'listenTcp' command as non-root ([#3292](https://github.com/edgexfoundry/edgex-go/issues/3292)) ([#5dc7e565](https://github.com/edgexfoundry/edgex-go/commits/5dc7e565))
- **security:** Rework Proxy Setup to use KongRoute struct for configuration ([#3228](https://github.com/edgexfoundry/edgex-go/issues/3228)) ([#2c126a99](https://github.com/edgexfoundry/edgex-go/commits/2c126a99))
    ```
    BREAKING CHANGE:
    Names for Route configuration has changed
    ```
- **all:** Modify config.Clients to use service key ([#afe0876a](https://github.com/edgexfoundry/edgex-go/commits/afe0876a))
- **security:** Update remaining SecretService references to be SecretStore ([#3189](https://github.com/edgexfoundry/edgex-go/issues/3189)) ([#afdb9f2a](https://github.com/edgexfoundry/edgex-go/commits/afdb9f2a))
- **all:** Remove support-logging and LoggingInfo config ([#2919](https://github.com/edgexfoundry/edgex-go/issues/2919)) ([#0163c92e](https://github.com/edgexfoundry/edgex-go/commits/0163c92e))
- **metadata:** Rename Put Command to Set Command ([#f6d4ba2d](https://github.com/edgexfoundry/edgex-go/commits/f6d4ba2d))
- **metadata:** Refactor device service update operation to DBClient ([#6b9e3f1f](https://github.com/edgexfoundry/edgex-go/commits/6b9e3f1f))
- **metadata:** Refactor provision watcher update operation to DBClient ([#0a10fb90](https://github.com/edgexfoundry/edgex-go/commits/0a10fb90))
- **metadata:** Refactor device profile update operation ([#f5f43703](https://github.com/edgexfoundry/edgex-go/commits/f5f43703))
- **metadata:** Refactor device update operation to DBClient ([#3059](https://github.com/edgexfoundry/edgex-go/issues/3059)) ([#abd2591c](https://github.com/edgexfoundry/edgex-go/commits/abd2591c))
- **metadata:** Rename PropertyValue's Type field to ValueType ([#7e47c43b](https://github.com/edgexfoundry/edgex-go/commits/7e47c43b))
- **metadata:** Remove all the Batch, DeleteByID, GetById API ([#15391329](https://github.com/edgexfoundry/edgex-go/commits/15391329))
- **metadata:** Move Transform func to go-mode-core-contract ([#e333ecd1](https://github.com/edgexfoundry/edgex-go/commits/e333ecd1))
- **notifications:** Rework of sending notifications ([#d2fe8064](https://github.com/edgexfoundry/edgex-go/commits/d2fe8064))
- **notifications:** Move ChannelSender interface to channel package ([#28350066](https://github.com/edgexfoundry/edgex-go/commits/28350066))
- **scheduler:** Remove runOnce feature ([#3549](https://github.com/edgexfoundry/edgex-go/issues/3549)) ([#5e3333aa](https://github.com/edgexfoundry/edgex-go/commits/5e3333aa))
- **sma:** Refactor sys-mgmt-executor ([#3543](https://github.com/edgexfoundry/edgex-go/issues/3543)) ([#f108a847](https://github.com/edgexfoundry/edgex-go/commits/f108a847))
- **sma:** Remove obsolete SMA v1 code ([#b4ce8a0d](https://github.com/edgexfoundry/edgex-go/commits/b4ce8a0d))
- **sma:** Remove unused configs and example ([#5120e818](https://github.com/edgexfoundry/edgex-go/commits/5120e818))

<a name="v1.3.1"></a>
## [v1.3.1] - 2021-02-08
### Features ✨
- **metadata:** Add service callback for deviceService AdminState Update API ([#a9476202](https://github.com/edgexfoundry/edgex-go/commits/a9476202))
### Bug Fixes 🐛
- Fix nil pointer error when update the unreachable DS adminState ([#c117ee17](https://github.com/edgexfoundry/edgex-go/commits/c117ee17))
- Upgrade to go-mod-messaging with ZMQ fix for Hanoi ([#3084](https://github.com/edgexfoundry/edgex-go/issues/3084)) ([#9a6eedb9](https://github.com/edgexfoundry/edgex-go/commits/9a6eedb9))
- **snap:** Fix redis snapshots ([#3102](https://github.com/edgexfoundry/edgex-go/issues/3102)) ([#12a188d7](https://github.com/edgexfoundry/edgex-go/commits/12a188d7))

<a name="v1.3.0"></a>
## [v1.3.0] - 2020-11-18
### Features ✨
- **all:** Add config setting for value used for ListenAndServe ([#2629](https://github.com/edgexfoundry/edgex-go/issues/2629)) ([#d3bef6b2](https://github.com/edgexfoundry/edgex-go/commits/d3bef6b2))
- **core-data:** Updated the Tags type to by object and added example to show how the data is represented in JSON. ([#212e9527](https://github.com/edgexfoundry/edgex-go/commits/212e9527))
- **core-data:** Add persisting of new Tags property on V1 & V2 Event models for Redis ([#2677](https://github.com/edgexfoundry/edgex-go/issues/2677)) ([#ae7f6d9e](https://github.com/edgexfoundry/edgex-go/commits/ae7f6d9e))
- **security:** Implement pluggable password generator ([#2659](https://github.com/edgexfoundry/edgex-go/issues/2659)) ([#ff532ada](https://github.com/edgexfoundry/edgex-go/commits/ff532ada))
- **core-data:** Add Tags property to Event in V1 & V2 swagger. ([#116c3839](https://github.com/edgexfoundry/edgex-go/commits/116c3839))
- **V2:** Add correlation id into log ([#16bfafab](https://github.com/edgexfoundry/edgex-go/commits/16bfafab))
- **core-data:** Event ID has to be pre-populated ([#2695](https://github.com/edgexfoundry/edgex-go/issues/2695)) ([#470d1768](https://github.com/edgexfoundry/edgex-go/commits/470d1768))
151af978))
- **metadata:** Optimize the error handling for deletion API ([#567a6ee1](https://github.com/edgexfoundry/edgex-go/commits/567a6ee1))
- **sdk:** Adding vault configuration default env variable ([#2673](https://github.com/edgexfoundry/edgex-go/issues/2673)) ([#1421448a](https://github.com/edgexfoundry/edgex-go/commits/1421448a))
- **security:** Implement encryption of vault master key ([#2574](https://github.com/edgexfoundry/edgex-go/issues/2574)) ([#09ff485f](https://github.com/edgexfoundry/edgex-go/commits/09ff485f))
- **security:** Add security-redis-bootstrap service ([#1a6876e5](https://github.com/edgexfoundry/edgex-go/commits/1a6876e5))
- **support-notifications:** Notification content type and long line ([#2699](https://github.com/edgexfoundry/edgex-go/issues/2699)) ([#855c38c3](https://github.com/edgexfoundry/edgex-go/commits/855c38c3))
### Snap
- **all:** Remove mongod ([#3cc3be18](https://github.com/edgexfoundry/edgex-go/commits/3cc3be18))
- **rules-engine:** Remove support-rulesengine ([#f881f5c4](https://github.com/edgexfoundry/edgex-go/commits/f881f5c4))
### Bug Fixes 🐛
- Use DB credentials for Redis Streams MesssageBus connection ([#2792](https://github.com/edgexfoundry/edgex-go/issues/2792)) ([#8ed4663e](https://github.com/edgexfoundry/edgex-go/commits/8ed4663e))
- Query event API w/ limit always returns first $n records (redis) ([#235aec4e](https://github.com/edgexfoundry/edgex-go/commits/235aec4e))
- Created timestamp is 0 on message queue ([#793f45a3](https://github.com/edgexfoundry/edgex-go/commits/793f45a3))
- ADD_PROXY_ROUTE fails if URL contains dot ([#6e12203f](https://github.com/edgexfoundry/edgex-go/commits/6e12203f))
- Fix path dependency in tokenprovider_linux_test.go ([#2641](https://github.com/edgexfoundry/edgex-go/issues/2641)) ([#04784571](https://github.com/edgexfoundry/edgex-go/commits/04784571))
- Allow startup duration/interval to be overridden via env vars ([#2649](https://github.com/edgexfoundry/edgex-go/issues/2649)) ([#b6e84d11](https://github.com/edgexfoundry/edgex-go/commits/b6e84d11))
- Use Itoa() instead of string() for int conversion ([#2663](https://github.com/edgexfoundry/edgex-go/issues/2663)) ([#6df8530f](https://github.com/edgexfoundry/edgex-go/commits/6df8530f))
- Get deviceProfile by ID when updating the valuedescriptor Should query device profile by name and id to prevent item not found error ([#234ed2e8](https://github.com/edgexfoundry/edgex-go/commits/234ed2e8))
- **data:** Modify the log level of event ([#2833](https://github.com/edgexfoundry/edgex-go/issues/2833)) ([#a54f4bf5](https://github.com/edgexfoundry/edgex-go/commits/a54f4bf5))
commits/471572d2))
- **metadata:** Refactor deviceProfile JSON and YAML POST API ([#2597](https://github.com/edgexfoundry/edgex-go/issues/2597)) ([#9098740b](https://github.com/edgexfoundry/edgex-go/commits/9098740b))
- **metadata:** Notify both device services when a Device is moved from one to the other ([#2716](https://github.com/edgexfoundry/edgex-go/issues/2716)) ([#bea4f5e6](https://github.com/edgexfoundry/edgex-go/commits/bea4f5e6))
- **metadata:** Device profile post returns 409 if id exists ([#172f3e63](https://github.com/edgexfoundry/edgex-go/commits/172f3e63))
- **notifications:** include From/To in SMTP header ([#2758](https://github.com/edgexfoundry/edgex-go/issues/2758)) ([#b3e2acdd](https://github.com/edgexfoundry/edgex-go/commits/b3e2acdd))
- **snap:** Disable asc version check ([#92e33c6b](https://github.com/edgexfoundry/edgex-go/commits/92e33c6b))
- **snap:** Update snap to use kong deb from bintray ([#335fa3dd](https://github.com/edgexfoundry/edgex-go/commits/335fa3dd))
- **snap:** Strip postgresql man pages ([#8a15cd27](https://github.com/edgexfoundry/edgex-go/commits/8a15cd27))
- **snap:** Strip commit+date from version ([#75c89412](https://github.com/edgexfoundry/edgex-go/commits/75c89412))
- **snap:** Remove external symlink to openresty ([#54f1720a](https://github.com/edgexfoundry/edgex-go/commits/54f1720a))
- **snap:** Remove support-logging ([#f3e829cf](https://github.com/edgexfoundry/edgex-go/commits/f3e829cf))
### Code Refactoring ♻
- Removed client monitoring ([#2595](https://github.com/edgexfoundry/edgex-go/issues/2595)) ([#ad8ce46e](https://github.com/edgexfoundry/edgex-go/commits/ad8ce46e))
### Other changes
- Remove security services initialization for mongodb ([#2567](https://github.com/edgexfoundry/edgex-go/issues/2567)) ([#80cc2cf8](https://github.com/edgexfoundry/edgex-go/commits/80cc2cf8))

<a name="v1.2.1"></a>
## [v1.2.1] - 2020-06-12
### Features ✨
- Add default MQTT optional MessageQueue values to enable env overrides ([#2564](https://github.com/edgexfoundry/edgex-go/issues/2564)) ([#e91925a3](https://github.com/edgexfoundry/edgex-go/commits/e91925a3))
### Bug Fixes 🐛
- Don't use hostname for webserver ListenAndServe ([#2579](https://github.com/edgexfoundry/edgex-go/issues/2579)) ([#525c6541](https://github.com/edgexfoundry/edgex-go/commits/525c6541))
- Fix: Allow overrides that have empty/blank value ([#3ccad16a](https://github.com/edgexfoundry/edgex-go/commits/3ccad16a))
- Added setting the Reading ID in the Events collection. ([#2575](https://github.com/edgexfoundry/edgex-go/issues/2575)) ([#fed02ba9](https://github.com/edgexfoundry/edgex-go/commits/fed02ba9))
- Accurately represent default port w/ EXPOSE in dockerfiles ([#2502f83b](https://github.com/edgexfoundry/edgex-go/commits/2502f83b))
- Missing fmt.Sprintf() in debug logging statement ([#4b30bbc4](https://github.com/edgexfoundry/edgex-go/commits/4b30bbc4))

<a name="v1.2.0"></a>
## [v1.2.0] - 2020-05-14
### Scheduler
- Remove QueueClient global and refactor its code ([#98dddcf2](https://github.com/edgexfoundry/edgex-go/commits/98dddcf2))
### Command
- Refactor to remove configuration global variable ([#2118](https://github.com/edgexfoundry/edgex-go/issues/2118)) ([#7aeef728](https://github.com/edgexfoundry/edgex-go/commits/7aeef728))
### Many
- Support new edgex-go security services ([#f09a2eaf](https://github.com/edgexfoundry/edgex-go/commits/f09a2eaf))
### Doc
- Save to docker-compose.yml ([#2040](https://github.com/edgexfoundry/edgex-go/issues/2040)) ([#8c7ea581](https://github.com/edgexfoundry/edgex-go/commits/8c7ea581))
### Feature
- **environment:** Allow uppercase environment overrides ([#14cb1f3e](https://github.com/edgexfoundry/edgex-go/commits/14cb1f3e))
### Security
- Fix non-empty token-provider Logging.File ([#2499](https://github.com/edgexfoundry/edgex-go/issues/2499)) ([#fdb80726](https://github.com/edgexfoundry/edgex-go/commits/fdb80726))
### Snap
- Allow SMA to be enabled/disabled ([#720bb04e](https://github.com/edgexfoundry/edgex-go/commits/720bb04e))
- Add Kuiper support ([#e57d4e41](https://github.com/edgexfoundry/edgex-go/commits/e57d4e41))
- Update db provider configure logic ([#bb82c305](https://github.com/edgexfoundry/edgex-go/commits/bb82c305))
- Include device-virtual binary dev profile ([#44f8e65f](https://github.com/edgexfoundry/edgex-go/commits/44f8e65f))
- Disable sys-mgmt-agent by default ([#d23fa061](https://github.com/edgexfoundry/edgex-go/commits/d23fa061))
- Enable redis security ([#675fad69](https://github.com/edgexfoundry/edgex-go/commits/675fad69))
- Enforce postgresql password auth ([#9bde2db7](https://github.com/edgexfoundry/edgex-go/commits/9bde2db7))
- Use per-service env overrides ([#7f63a8d3](https://github.com/edgexfoundry/edgex-go/commits/7f63a8d3))
- Update default db to be redis ([#e1cef487](https://github.com/edgexfoundry/edgex-go/commits/e1cef487))
- Remove device-random ([#67bc4086](https://github.com/edgexfoundry/edgex-go/commits/67bc4086))
### Notifications
- Refactor to remove Configuration global variable ([#c021313d](https://github.com/edgexfoundry/edgex-go/commits/c021313d))
- Refactor to remove dbClient global variable ([#5f01098a](https://github.com/edgexfoundry/edgex-go/commits/5f01098a))
- Refactor to remove LoggingClient global variable ([#e37ee154](https://github.com/edgexfoundry/edgex-go/commits/e37ee154))
### Bug Fixes 🐛
- Add Redis connection test during client creation so error will trigger retry ([#8dfb5d32](https://github.com/edgexfoundry/edgex-go/commits/8dfb5d32))
- Update to use go-mod-bootstrap to fix issue with override un-done. ([#2536](https://github.com/edgexfoundry/edgex-go/issues/2536)) ([#ac53844b](https://github.com/edgexfoundry/edgex-go/commits/ac53844b))
- Add generation of application-service vault token for shared DB credentials ([#af1eaf2f](https://github.com/edgexfoundry/edgex-go/commits/af1eaf2f))
- Add call to  Message Bus Connect() ([#2467](https://github.com/edgexfoundry/edgex-go/issues/2467)) ([#2cabbc24](https://github.com/edgexfoundry/edgex-go/commits/2cabbc24))
- [#2034](https://github.com/edgexfoundry/edgex-go/issues/2034) fixes bug around named return values ([#dce4ecfd](https://github.com/edgexfoundry/edgex-go/commits/dce4ecfd))
### Code Refactoring ♻

<a name="v1.1.0"></a>
## [v1.1.0] - 2019-11-14
### Features ✨
- **config-seed:** Change Config Seed rules engine properties so messages are received from App-Service-Configurable ([#dd6fb282](https://github.com/edgexfoundry/edgex-go/commits/dd6fb282))
### Feature
- **query-params:** Pass QueryParams through EdgeX to Device Services ([#1571](https://github.com/edgexfoundry/edgex-go/issues/1571)) ([#4d7ed080](https://github.com/edgexfoundry/edgex-go/commits/4d7ed080))
### Bug Fixes 🐛
- [#2034](https://github.com/edgexfoundry/edgex-go/issues/2034) fixes bug around named return values ([#45cdcb29](https://github.com/edgexfoundry/edgex-go/commits/45cdcb29))
- **config-seed:** Slice bound out of range on Windows ([#1606](https://github.com/edgexfoundry/edgex-go/issues/1606)) ([#7ee64677](https://github.com/edgexfoundry/edgex-go/commits/7ee64677))

<a name="v1.0.0"></a>
## [v1.0.0] - 2019-06-25
### Many
- Rename ReadMaxLimit to MaxResultCount, set default to 50k ([#499cd073](https://github.com/edgexfoundry/edgex-go/commits/499cd073))
### FIX
- Client monitor update in milliseconds, not seconds ([#cd852482](https://github.com/edgexfoundry/edgex-go/commits/cd852482))
- Event ids blank when exported ([#f9b26649](https://github.com/edgexfoundry/edgex-go/commits/f9b26649))

<a name="0.7.1"></a>
## [0.7.1] - 2018-12-10
### FIX
- Client monitor update in milliseconds, not seconds ([#7424180a](https://github.com/edgexfoundry/edgex-go/commits/7424180a))

<a name="0.7.0"></a>
## [0.7.0] - 2018-11-16
### BUG
- Consul values overridden at service start ([#b5d54ea5](https://github.com/edgexfoundry/edgex-go/commits/b5d54ea5))
### Snap
- Move bin and config dirs into snap/local/ ([#aada7c16](https://github.com/edgexfoundry/edgex-go/commits/aada7c16))
### Fix
- LogLevel field name in JSON, criteria in Mongo ([#7534e412](https://github.com/edgexfoundry/edgex-go/commits/7534e412))
### Metadata
- Check that db type is mongo before getting a session ([#410d0046](https://github.com/edgexfoundry/edgex-go/commits/410d0046))

<a name="v0.0.0"></a>
## v0.0.0 - 2021-02-01
### Features ✨
- Add Tags property to Event in V1 & V2 swagger. ([#116c3839](https://github.com/edgexfoundry/edgex-go/commits/116c3839))
- Add default MQTT optional MessageQueue values to enable env overrides ([#2564](https://github.com/edgexfoundry/edgex-go/issues/2564)) ([#e91925a3](https://github.com/edgexfoundry/edgex-go/commits/e91925a3))
- Add persisting of new Tags property on V1 & V2 Event models for Redis ([#2677](https://github.com/edgexfoundry/edgex-go/issues/2677)) ([#ae7f6d9e](https://github.com/edgexfoundry/edgex-go/commits/ae7f6d9e))
- Implement pluggable password generator ([#2659](https://github.com/edgexfoundry/edgex-go/issues/2659)) ([#ff532ada](https://github.com/edgexfoundry/edgex-go/commits/ff532ada))
- Add config setting for value used for ListenAndServe ([#2629](https://github.com/edgexfoundry/edgex-go/issues/2629)) ([#d3bef6b2](https://github.com/edgexfoundry/edgex-go/commits/d3bef6b2))
- Updated the Tags type to by object and added example to show how the data is represented in JSON. ([#212e9527](https://github.com/edgexfoundry/edgex-go/commits/212e9527))
- **V2:** Add correlation id into log ([#16bfafab](https://github.com/edgexfoundry/edgex-go/commits/16bfafab))
- **config-seed:** Change Config Seed rules engine properties so messages are received from App-Service-Configurable ([#dd6fb282](https://github.com/edgexfoundry/edgex-go/commits/dd6fb282))
- **core-data:** Event ID has to be pre-populated ([#2695](https://github.com/edgexfoundry/edgex-go/issues/2695)) ([#470d1768](https://github.com/edgexfoundry/edgex-go/commits/470d1768))
- **metadata:** Optimize the error handling for deletion API ([#567a6ee1](https://github.com/edgexfoundry/edgex-go/commits/567a6ee1))
- **metadata:** Add service callback for deviceService AdminState Update API ([#a9476202](https://github.com/edgexfoundry/edgex-go/commits/a9476202))
- **sdk:** Adding vault configuration default env variable ([#2673](https://github.com/edgexfoundry/edgex-go/issues/2673)) ([#1421448a](https://github.com/edgexfoundry/edgex-go/commits/1421448a))
- **security:** Implement encryption of vault master key ([#2574](https://github.com/edgexfoundry/edgex-go/issues/2574)) ([#09ff485f](https://github.com/edgexfoundry/edgex-go/commits/09ff485f))
- **security:** Add security-redis-bootstrap service ([#1a6876e5](https://github.com/edgexfoundry/edgex-go/commits/1a6876e5))
- **support-notifications:** notification content type and long line ([#2699](https://github.com/edgexfoundry/edgex-go/issues/2699)) ([#855c38c3](https://github.com/edgexfoundry/edgex-go/commits/855c38c3))
### Core
- Unified core and metadata db interfaces package name ([#9e847c16](https://github.com/edgexfoundry/edgex-go/commits/9e847c16))
- Create a new package for db access ([#50d46abf](https://github.com/edgexfoundry/edgex-go/commits/50d46abf))
### Fix
- LogLevel field name in JSON, criteria in Mongo ([#7534e412](https://github.com/edgexfoundry/edgex-go/commits/7534e412))
### BUG
- Consul values overridden at service start ([#b5d54ea5](https://github.com/edgexfoundry/edgex-go/commits/b5d54ea5))
### Snap
- Remove mongod ([#3cc3be18](https://github.com/edgexfoundry/edgex-go/commits/3cc3be18))
- Remove support-rulesengine ([#f881f5c4](https://github.com/edgexfoundry/edgex-go/commits/f881f5c4))
- Allow SMA to be enabled/disabled ([#720bb04e](https://github.com/edgexfoundry/edgex-go/commits/720bb04e))
- Add Kuiper support ([#e57d4e41](https://github.com/edgexfoundry/edgex-go/commits/e57d4e41))
- Update db provider configure logic ([#bb82c305](https://github.com/edgexfoundry/edgex-go/commits/bb82c305))
- Include device-virtual binary dev profile ([#44f8e65f](https://github.com/edgexfoundry/edgex-go/commits/44f8e65f))
- Disable sys-mgmt-agent by default ([#d23fa061](https://github.com/edgexfoundry/edgex-go/commits/d23fa061))
- Enable redis security ([#675fad69](https://github.com/edgexfoundry/edgex-go/commits/675fad69))
- Enforce postgresql password auth ([#9bde2db7](https://github.com/edgexfoundry/edgex-go/commits/9bde2db7))
- Use per-service env overrides ([#7f63a8d3](https://github.com/edgexfoundry/edgex-go/commits/7f63a8d3))
- Fix secretstore-setup's token-provider ([#a5387499](https://github.com/edgexfoundry/edgex-go/commits/a5387499))
- Apply gosu fix from fuji ([#8ec53dd5](https://github.com/edgexfoundry/edgex-go/commits/8ec53dd5))
- Fix config-seed startup ([#49731fc6](https://github.com/edgexfoundry/edgex-go/commits/49731fc6))
- Fix device-random startup ([#f6c70a3a](https://github.com/edgexfoundry/edgex-go/commits/f6c70a3a))
- Fix sys-mgmt-agent executor ([#9c28a602](https://github.com/edgexfoundry/edgex-go/commits/9c28a602))
- Update pre-refresh hook for geneva ([#9b43f8a1](https://github.com/edgexfoundry/edgex-go/commits/9b43f8a1))
- Update epoch to 3 for geneva ([#353e084e](https://github.com/edgexfoundry/edgex-go/commits/353e084e))
- Remove device-random ([#67bc4086](https://github.com/edgexfoundry/edgex-go/commits/67bc4086))
- Move bin and config dirs into snap/local/ ([#aada7c16](https://github.com/edgexfoundry/edgex-go/commits/aada7c16))
### FIX
- Client monitor update in milliseconds, not seconds ([#cd852482](https://github.com/edgexfoundry/edgex-go/commits/cd852482))
- Event ids blank when exported ([#f9b26649](https://github.com/edgexfoundry/edgex-go/commits/f9b26649))
### Feature
- **environment:** Allow uppercase environment overrides ([#14cb1f3e](https://github.com/edgexfoundry/edgex-go/commits/14cb1f3e))
- **query-params:** Pass QueryParams through EdgeX to Device Services ([#1571](https://github.com/edgexfoundry/edgex-go/issues/1571)) ([#4d7ed080](https://github.com/edgexfoundry/edgex-go/commits/4d7ed080))
### Security
- Fix non-empty token-provider Logging.File ([#2499](https://github.com/edgexfoundry/edgex-go/issues/2499)) ([#fdb80726](https://github.com/edgexfoundry/edgex-go/commits/fdb80726))
### Refact
- Use latest go-mod-bootstrap with self seeding, remove config-seed & remove Docker profiles ([#28c25972](https://github.com/edgexfoundry/edgex-go/commits/28c25972))
### Scheduler
- Remove QueueClient global and refactor its code ([#98dddcf2](https://github.com/edgexfoundry/edgex-go/commits/98dddcf2))
### Many
- Support new edgex-go security services ([#f09a2eaf](https://github.com/edgexfoundry/edgex-go/commits/f09a2eaf))
- Rename vault-config.json to vault-config.hcl ([#6d2924b2](https://github.com/edgexfoundry/edgex-go/commits/6d2924b2))
- Rename ReadMaxLimit to MaxResultCount, set default to 50k ([#499cd073](https://github.com/edgexfoundry/edgex-go/commits/499cd073))
### Command
- Refactor to remove configuration global variable ([#2118](https://github.com/edgexfoundry/edgex-go/issues/2118)) ([#7aeef728](https://github.com/edgexfoundry/edgex-go/commits/7aeef728))
### Notifications
- Refactor to remove Configuration global variable ([#c021313d](https://github.com/edgexfoundry/edgex-go/commits/c021313d))
- Refactor to remove dbClient global variable ([#5f01098a](https://github.com/edgexfoundry/edgex-go/commits/5f01098a))
- Refactor to remove LoggingClient global variable ([#e37ee154](https://github.com/edgexfoundry/edgex-go/commits/e37ee154))
### Metadata
- Check that db type is mongo before getting a session ([#410d0046](https://github.com/edgexfoundry/edgex-go/commits/410d0046))
### Bug Fixes 🐛
- Upgrade to go-mod-messaging with ZMQ fix for Hanoi ([#3084](https://github.com/edgexfoundry/edgex-go/issues/3084)) ([#9a6eedb9](https://github.com/edgexfoundry/edgex-go/commits/9a6eedb9))
- [#2034](https://github.com/edgexfoundry/edgex-go/issues/2034) fixes bug around named return values ([#dce4ecfd](https://github.com/edgexfoundry/edgex-go/commits/dce4ecfd))
- Add call to  Message Bus Connect() ([#2467](https://github.com/edgexfoundry/edgex-go/issues/2467)) ([#2cabbc24](https://github.com/edgexfoundry/edgex-go/commits/2cabbc24))
- Fix nil pointer error when update the unreachable DS adminState ([#c117ee17](https://github.com/edgexfoundry/edgex-go/commits/c117ee17))
- Add generation of application-service vault token for shared DB credentials ([#af1eaf2f](https://github.com/edgexfoundry/edgex-go/commits/af1eaf2f))
- Added setting the Reading ID in the Events collection. ([#2575](https://github.com/edgexfoundry/edgex-go/issues/2575)) ([#fed02ba9](https://github.com/edgexfoundry/edgex-go/commits/fed02ba9))
- Fix: Allow overrides that have empty/blank value ([#3ccad16a](https://github.com/edgexfoundry/edgex-go/commits/3ccad16a))
- Use DB credentials for Redis Streams MesssageBus connection ([#2792](https://github.com/edgexfoundry/edgex-go/issues/2792)) ([#8ed4663e](https://github.com/edgexfoundry/edgex-go/commits/8ed4663e))
- Don't use hostname for webserver ListenAndServe ([#2579](https://github.com/edgexfoundry/edgex-go/issues/2579)) ([#525c6541](https://github.com/edgexfoundry/edgex-go/commits/525c6541))
- Query event API w/ limit always returns first $n records (redis) ([#235aec4e](https://github.com/edgexfoundry/edgex-go/commits/235aec4e))
- Created timestamp is 0 on message queue ([#793f45a3](https://github.com/edgexfoundry/edgex-go/commits/793f45a3))
- Allow startup duration/interval to be overridden via env vars ([#2649](https://github.com/edgexfoundry/edgex-go/issues/2649)) ([#b6e84d11](https://github.com/edgexfoundry/edgex-go/commits/b6e84d11))
- ADD_PROXY_ROUTE fails if URL contains dot ([#6e12203f](https://github.com/edgexfoundry/edgex-go/commits/6e12203f))
- Fix path dependency in tokenprovider_linux_test.go ([#2641](https://github.com/edgexfoundry/edgex-go/issues/2641)) ([#04784571](https://github.com/edgexfoundry/edgex-go/commits/04784571))
- Use Itoa() instead of string() for int conversion ([#2663](https://github.com/edgexfoundry/edgex-go/issues/2663)) ([#6df8530f](https://github.com/edgexfoundry/edgex-go/commits/6df8530f))
- Get deviceProfile by ID when updating the valuedescriptor Should query device profile by name and id to prevent item not found error ([#234ed2e8](https://github.com/edgexfoundry/edgex-go/commits/234ed2e8))
- **config-seed:** Slice bound out of range on Windows ([#1606](https://github.com/edgexfoundry/edgex-go/issues/1606)) ([#7ee64677](https://github.com/edgexfoundry/edgex-go/commits/7ee64677))
- **data:** Modify the log level of event ([#2833](https://github.com/edgexfoundry/edgex-go/issues/2833)) ([#a54f4bf5](https://github.com/edgexfoundry/edgex-go/commits/a54f4bf5))
- **metadata:** Device PATCH V2 API should check service and profile ([#2862](https://github.com/edgexfoundry/edgex-go/issues/2862)) ([#471572d2](https://github.com/edgexfoundry/edgex-go/commits/471572d2))
- **metadata:** Refactor deviceProfile JSON and YAML POST API ([#2597](https://github.com/edgexfoundry/edgex-go/issues/2597)) ([#9098740b](https://github.com/edgexfoundry/edgex-go/commits/9098740b))
- **metadata:** V2 GET /deviceservice/all returns inconsistent response when specifying labels or not ([#08b8cf9d](https://github.com/edgexfoundry/edgex-go/commits/08b8cf9d))
- **metadata:** Device profile post returns 409 if id exists ([#172f3e63](https://github.com/edgexfoundry/edgex-go/commits/172f3e63))
- **metadata:** correct V2 parsing err response ([#c4d32136](https://github.com/edgexfoundry/edgex-go/commits/c4d32136))
- **metadata:** Notify both device services when a Device is moved from one to the other ([#2716](https://github.com/edgexfoundry/edgex-go/issues/2716)) ([#bea4f5e6](https://github.com/edgexfoundry/edgex-go/commits/bea4f5e6))
- **notifications:** Include From/To in SMTP header ([#2758](https://github.com/edgexfoundry/edgex-go/issues/2758)) ([#b3e2acdd](https://github.com/edgexfoundry/edgex-go/commits/b3e2acdd))
- **snap:** remove support-logging ([#f3e829cf](https://github.com/edgexfoundry/edgex-go/commits/f3e829cf))
### Code Refactoring ♻
- Removed client monitoring ([#2595](https://github.com/edgexfoundry/edgex-go/issues/2595)) ([#ad8ce46e](https://github.com/edgexfoundry/edgex-go/commits/ad8ce46e))
- **all:** Use constant for Redis key in V2 ([#df6ae563](https://github.com/edgexfoundry/edgex-go/commits/df6ae563))
- **core-data:** Error handling for V2 API ([#2681](https://github.com/edgexfoundry/edgex-go/issues/2681)) ([#79f01a0b](https://github.com/edgexfoundry/edgex-go/commits/79f01a0b))
### Documentation 📖
- Update ZMQ for module directory structure. ([#2191](https://github.com/edgexfoundry/edgex-go/issues/2191)) ([#32c2c55f](https://github.com/edgexfoundry/edgex-go/commits/32c2c55f))
- **all:** Add multiple responses schemas to V2 Swagger files ([#82e94d13](https://github.com/edgexfoundry/edgex-go/commits/82e94d13))
- **all:** Update response codes in V2 Swagger files ([#3130a5bf](https://github.com/edgexfoundry/edgex-go/commits/3130a5bf))
- **data:** Update examples in V2 API Swagger file ([#9a9f8dfa](https://github.com/edgexfoundry/edgex-go/commits/9a9f8dfa))
### Other changes
- Remove security services initialization for mongodb ([#2567](https://github.com/edgexfoundry/edgex-go/issues/2567)) ([#80cc2cf8](https://github.com/edgexfoundry/edgex-go/commits/80cc2cf8))
