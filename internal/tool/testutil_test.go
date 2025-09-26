/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tool

import (
	"net"
	"net/url"
)

// getLocalhostServerURL extracts the port from a test server URL and returns a localhost URL.
func getLocalhostServerURL(serverURL string) string {
	u, err := url.Parse(serverURL)
	if err != nil {
		panic("invalid server URL: " + err.Error())
	}
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		panic("invalid host:port in server URL: " + err.Error())
	}
	return "localhost:" + port
}
