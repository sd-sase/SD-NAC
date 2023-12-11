const { SCOPE_INSERT, SCOPE_UPDATE, SCOPE_DELETE } = require('../config');
const collection_url = '/configuration/security_events';
const resource_url = id => `/configuration/security_event/${id}`;
const fixture = 'collections/securityEvent.json';
const timeout = 10E3;

module.exports = {
  id: 'securityEvents',
  description: 'Security Events',
  tests: [
    {
      description: 'Security Events - Create New',
      scope: SCOPE_INSERT,
      url: collection_url,
      fixture,
      timeout,
      selectors: {
        buttonNewSelectors: ['button[type="button"]:contains(New)'],
      },
      interceptors: [
        {
          method: 'POST',
          url: '/api/**/config/security_events',
          expectRequest: (request, fixture) => {
            Object.keys(fixture).forEach(key => {
              expect(request.body).to.have.property(key)
              expect(request.body[key]).to.deep.equal(fixture[key], key)
            })
          },
          expectResponse: (response, fixture) => {
            expect(response.statusCode).to.equal(201)
            const { body: { id } = {} } = response
            return { id } // push `id` to fixture
          }
        }
      ]
    },
    {
      description: 'Security Events - Update Existing',
      scope: SCOPE_UPDATE,
      idFrom: (_, cache) => cache.id,
      fixture,
      timeout,
      url: resource_url,
      interceptors: [
        {
          method: '+(PATCH|PUT)',
          url: '/api/**/config/security_event/**',
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
      description: 'Security Events - Delete Existing',
      scope: SCOPE_DELETE,
      idFrom: (_, cache) => cache.id,
      fixture,
      timeout,
      url: resource_url,
      interceptors: [
        {
          method: 'DELETE', url: '/api/**/config/security_event/**', expectResponse: (response, fixture) => {
            expect(response.statusCode).to.equal(200)
          }
        }
      ]
    }
  ]
};