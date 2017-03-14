import React from 'react'
import { shallow } from 'enzyme'

import Error from '../../src/components/Error.jsx'
import { RequestError, InternalError, ApiError } from 'redux-api-middleware'

function setup(error) {
  return shallow(<Error error={error} />)
}

describe('Error', () => {
  it('should render generic error', () => {
    const enzymeWrapper = setup(new Error())
    expect(enzymeWrapper.find('div').hasClass('error')).toBe(true)
    expect(enzymeWrapper.find('div').text()).toEqual('An unknown error occurred. Please contact your friendly QuotaService owners for help.')
  })

  it('should render InternalError', () => {
    const enzymeWrapper = setup(new InternalError())
    expect(enzymeWrapper.find('div').hasClass('error')).toBe(true)
    expect(enzymeWrapper.find('div').text()).toEqual('An unknown error occurred. Please contact your friendly QuotaService owners for help.')
  })

  it('should render ApiError', () => {
    let enzymeWrapper = setup(new ApiError(400, 'bad request'))
    expect(enzymeWrapper.find('div').hasClass('error')).toBe(true)
    expect(enzymeWrapper.find('div').text()).toEqual('An error occurred: "bad request"')

    enzymeWrapper = setup(new ApiError(400, 'bad request', { description: 'This is the description.' }))
    expect(enzymeWrapper.find('div').hasClass('error')).toBe(true)
    expect(enzymeWrapper.find('div').text()).toEqual('This is the description.')
  })

  it('should render RequestError', () => {
    let enzymeWrapper = setup(new RequestError('network timeout'))
    expect(enzymeWrapper.find('div').hasClass('error')).toBe(true)
    expect(enzymeWrapper.find('div').text()).toEqual('A network error occurred: "network timeout"')
  })
})
