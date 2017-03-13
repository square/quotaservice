import React from 'react'
import { shallow } from 'enzyme'

import Change from '../../src/components/Change.jsx'
import * as actions from '../../src/actions/mutable.jsx'

function setup(change) {
  const props = {
    className: "present",
    change: change
  }

  return shallow(<Change {...props} />)
}


describe('Change', () => {
  it('should render ADD_NAMESPACE text properly', () => {
    const enzymeWrapper = setup({
      type: actions.ADD_NAMESPACE,
      key: "test.namespace"
    })

    expect(enzymeWrapper.find('div').hasClass('change present')).toBe(true)
    expect(enzymeWrapper.find('span').hasClass('change-text')).toBe(true)
    expect(enzymeWrapper.find('span').text()).toBe("add test.namespace")
  })

  it('should render ADD_BUCKET text properly', () => {
    const enzymeWrapper = setup({
      type: actions.ADD_BUCKET,
      key: "test.namespace.bucket"
    })

    expect(enzymeWrapper.find('div').hasClass('change present')).toBe(true)
    expect(enzymeWrapper.find('span').hasClass('change-text')).toBe(true)
    expect(enzymeWrapper.find('span').text()).toBe("add test.namespace.bucket")
  })

  it('should render UPDATE_NAMESPACE text properly', () => {
    const enzymeWrapper = setup({
      type: actions.UPDATE_NAMESPACE,
      key: "test.namespace.foo",
      value: "bar"
    })

    expect(enzymeWrapper.find('div').hasClass('change present')).toBe(true)
    expect(enzymeWrapper.find('span').hasClass('change-text')).toBe(true)
    expect(enzymeWrapper.find('span').text()).toBe("set test.namespace.foo to \"bar\"")
  })

  it('should render UPDATE_BUCKET text properly', () => {
    const enzymeWrapper = setup({
      type: actions.UPDATE_BUCKET,
      key: "test.namespace.bucket.foo",
      value: "bar"
    })

    expect(enzymeWrapper.find('div').hasClass('change present')).toBe(true)
    expect(enzymeWrapper.find('span').hasClass('change-text')).toBe(true)
    expect(enzymeWrapper.find('span').text()).toBe("set test.namespace.bucket.foo to \"bar\"")
  })

  it('should render REMOVE_NAMESPACE text properly', () => {
    const enzymeWrapper = setup({
      type: actions.REMOVE_NAMESPACE,
      key: "test.namespace.foo",
    })

    expect(enzymeWrapper.find('div').hasClass('change present')).toBe(true)
    expect(enzymeWrapper.find('span').hasClass('change-text')).toBe(true)
    expect(enzymeWrapper.find('span').text()).toBe("remove test.namespace.foo")
  })

  it('should render UPDATE_BUCKET text properly', () => {
    const enzymeWrapper = setup({
      type: actions.REMOVE_BUCKET,
      key: "test.namespace.bucket"
    })

    expect(enzymeWrapper.find('div').hasClass('change present')).toBe(true)
    expect(enzymeWrapper.find('span').hasClass('change-text')).toBe(true)
    expect(enzymeWrapper.find('span').text()).toBe("remove test.namespace.bucket")
  })

  it('should render unknown change', () => {
    const enzymeWrapper = setup({
      type: "none"
    })

    expect(enzymeWrapper.find('div').text()).toBe(`Unknown change: ${JSON.stringify({ type: "none" })}`)
  })
})
