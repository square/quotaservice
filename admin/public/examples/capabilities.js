/**
 * This is an example of how capabilities could be implemented
 * via DOM event based hooks.
 */
(function() {
  let capabilities;

  // enable capabilities
  window.QUOTASERVICE_CAPABILITIES = true;

  function transformRawCapabilities(rawCapabilities) {
    const capabilities = {};

    Object.keys(rawCapabilities).forEach(namespaceName => {
      // splitting by : and taking first part allows complex names spaces to match application names
      namespaceName = namespaceName.split(/:/)[0];
      capabilities[namespaceName] =
        rawCapabilities[namespaceName].find(
          group => ['deployers', 'owners'].indexOf(group) !== -1
        ) !== undefined;
    });

    return capabilities;
  }

  window.addEventListener('QuotaService.fetchCapabilities', e => {
    capabilities = fetch('/api/capabilities')
      .then(response => response.json())
      .then(transformRawCapabilities);

    capabilities.then(e.detail.callback);
  });

  window.addEventListener('QuotaService.getCapabilities', e => {
    const { callback } = e.detail;
    let { namespaceName } = e.detail;

    if (!capabilities) {
      callback(false);
      return;
    }

    console.log(`Fetching capabiligies for "${namespaceName}"`);

    namespaceName = namespaceName.split(/:/)[0];

    capabilities.then(capabilities =>
      callback(capabilities[namespaceName] === true)
    );
  });
})();
