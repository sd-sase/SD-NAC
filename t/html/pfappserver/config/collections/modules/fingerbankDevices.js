const { SCOPE_INSERT, SCOPE_UPDATE, SCOPE_DELETE } = require('../config');

const collection_url = 'configuration/fingerbank/devices';
const resource_url = id => `/configuration/fingerbank/local/device/${id}`;
const fixture = 'collections/fingerbank/device.json';

module.exports = {
  id: 'fingerbankDevices',
  description: 'Fingerbank Devices',
  tests: [
    {
      description: 'Fingerbank Devices - Create New',
      scope: SCOPE_INSERT,
      fixture,
      url: collection_url,
      selectors: {
        buttonNewSelectors: ['button[type="button"]:contains(New Device)'],
      },
      interceptors: [
        {
          method: 'POST',
          url: '/api/**/fingerbank/local/devices',
          expectRequest: (request, fixture) => {
            Object.keys(fixture).forEach(key => {
              expect(request.body).to.have.property(key)
              expect(request.body[key]).to.deep.equal(fixture[key], key)
            })
          },
          expectResponse: (response, fixture) => {
            expect(response.statusCode).to.equal(200)
            return response.body // push `id` to fixture
          }
        }
      ]
    },
    {
      description: 'Fingerbank Devices - Update Existing',
      scope: SCOPE_UPDATE,
      fixture,
      url: resource_url,
      idFrom: (_, cache) => cache.id,
      interceptors: [
        {
          method: '+(PATCH|PUT)',
          url: '/api/**/fingerbank/local/device/**',
          expectRequest: (request, fixture) => {
            Object.keys(fixture).forEach(key => {
              expect(request.body).to.have.property(key)
              expect(request.body[key]).to.deep.equal(fixture[key], key)
            })
          },
          expectResponse: (response, fixture) => {
            expect(response.statusCode).to.equal(200)
          }
        }
      ]
    },
    {
      description: 'Fingerbank Devices - Delete Existing',
      scope: SCOPE_DELETE,
      fixture,
      url: resource_url,
      idFrom: (_, cache) => cache.id,
      interceptors: [
        {
          method: 'DELETE', url: '/api/**/fingerbank/local/device/**', expectResponse: (response, fixture) => {
            expect(response.statusCode).to.equal(200)
          }
        }
      ]
    }
  ]
};