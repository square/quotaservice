import Promise from '../../src/promise';
import { shallow } from 'enzyme';
import toJSON from 'enzyme-to-json';
import React from 'react';

import QuotaService from '../../src/containers/QuotaService.jsx';

describe('QuotaService', () => {
  let props;
  let fetchConfigsPromise;

  beforeEach(() => {
    fetchConfigsPromise = Promise.resolve({});

    props = {
      env: {
        capabilities: true,
      },
      actions: {
        fetchConfigs: () => fetchConfigsPromise,
        fetchCapabilities: jest.fn(),
      },
      dispatch: jest.fn(),
      namespaces: {},
      stats: {},
      configs: {},
      capabilities: {},
      currentVersion: 1,
    };
  });

  it('renders ok', () => {
    const component = shallow(<QuotaService {...props} />);
    const tree = toJSON(component);
    expect(tree).toMatchSnapshot();
  });

  describe('capabilities', () => {
    it('are requested', () => {
      shallow(<QuotaService {...props} />);
      return fetchConfigsPromise.then(() =>
        expect(props.actions.fetchCapabilities.mock.calls.length).toBe(1)
      );
    });

    it('are not requested', () => {
      props.env.capabilities = false;
      shallow(<QuotaService {...props} />);
      return fetchConfigsPromise.then(() =>
        expect(props.actions.fetchCapabilities.mock.calls.length).toBe(0)
      );
    });
  });
});
