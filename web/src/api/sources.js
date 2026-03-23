import request from '@/utils/request'

export function getSources() {
  return request({
    url: '/sources',
    method: 'get'
  })
}

export function getSource(id) {
  return request({
    url: `/sources/${id}`,
    method: 'get'
  })
}

export function createSource(data) {
  return request({
    url: '/sources',
    method: 'post',
    data
  })
}

export function updateSource(id, data) {
  return request({
    url: `/sources/${id}`,
    method: 'put',
    data
  })
}

export function deleteSource(id) {
  return request({
    url: `/sources/${id}`,
    method: 'delete'
  })
}

export function syncSource(id) {
  return request({
    url: `/sources/${id}/sync`,
    method: 'post'
  })
}

export function testSource(url) {
  return request({
    url: '/sources/test',
    method: 'post',
    data: { url }
  })
}
