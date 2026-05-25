/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package errorz

type ErrorHolder interface {
	GetError() error
	SetError(error) bool
	HasError() bool
}

type ErrorHolderImpl struct {
	Err error
}

func (holder *ErrorHolderImpl) GetError() error {
	return holder.Err
}

func (holder *ErrorHolderImpl) SetError(err error) bool {
	if err != nil && holder.Err == nil {
		holder.Err = err
	}
	return holder.Err != nil
}

func (holder *ErrorHolderImpl) HasError() bool {
	return holder.Err != nil
}
