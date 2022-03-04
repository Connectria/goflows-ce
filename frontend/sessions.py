import requests

class Context:

  tenantId = None

  def __init__(self, request):

    self.tenantId = None

    myTenantId = request.headers.get('tenantId')

    if myTenantId and myTenantId.isnumeric():

      self.tenantId = int(myTenantId)
