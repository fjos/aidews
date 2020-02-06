package policy

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestEqual(t *testing.T) {
	j0 := `{"a": "aye"}`
	j1 := `{"a": "aye"}`
	j2 := `{"a": "bee"}`
	eq, _ := Equal([]byte(j0), []byte(j1))
	if !eq {
		t.Error("equal json strings incorrectly appear unequal")
	}
	eq, _ = Equal([]byte(j0), []byte(j2))
	if eq {
		t.Error("unequal json strings incorrectly appear equal")
	}
}

func TestEqual_errUnmarshal(t *testing.T) {
	var err error
	j0 := `{"a": "aye"}`
	j1 := `{"a": 'bad"}`
	want := `invalid character '\'' looking for beginning of value`
	_, err = Equal([]byte(j0), []byte(j1))
	if err == nil {
		t.Errorf("no error; want: %s", want)
	} else {
		if err.Error() != want {
			t.Errorf("unexpected error; want: %s, got: %s", want, err)
		}
	}
	_, err = Equal([]byte(j1), []byte(j0))
	if err == nil {
		t.Errorf("no error; want: %s", want)
	} else {
		if err.Error() != want {
			t.Errorf("unexpected error; want: %s, got: %s", want, err)
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	res := &struct {
		SS0 StrOrSlice
		SS1 StrOrSlice
	}{
		SS0: StrOrSlice([]string{"test"}),
		SS1: StrOrSlice([]string{"test", "slice"}),
	}
	b, _ := json.Marshal(res)
	want := `{"SS0":"test","SS1":["test","slice"]}`
	if string(b) != want {
		t.Errorf("incorrectly marshaled; want: %s, got: %s", want, string(b))
	}
}

func TestUnmarshalJSON(t *testing.T) {
	res := struct {
		SS0 StrOrSlice
		SS1 StrOrSlice
	}{}
	j := []byte(`{"SS0": "test", "SS1": ["test", "slice"]}`)
	if err := json.Unmarshal(j, &res); err != nil {
		t.Error("error during Unmarshal")
	}
}

func TestUnmarshalJSON_errSliceUnmarshal(t *testing.T) {
	res := struct {
		SS0 StrOrSlice
		SS1 StrOrSlice
	}{}
	j := []byte(`{"SS0": "test", "SS1": {"noobjects": ["test", "slice"]}}`)
	want := "json: cannot unmarshal object into Go struct field .SS1 of type []string"
	if err := json.Unmarshal(j, &res); err == nil {
		t.Errorf("expected error; want: %s", err)
	} else {
		if err.Error() != want {
			t.Errorf("unexpected error; want: %s, got: %s", want, err)
		}
	}
}

func TestIAMPolicyJSON(t *testing.T) {
	policy := IAMPolicy{
		Version: "12",
	}
	out, err := json.Marshal(policy)
	if err != nil {
		t.Errorf("unexpected error; got: %s", err)
	}
	want := "{\"Version\":\"12\",\"Statement\":null}"
	if string(out) != want {
		t.Errorf("unexpected out; want: %s got %s", want, string(out))
	}
}

func TestIAMPolicyStatementEmptyJSON(t *testing.T) {
	policy := IAMPolicyStatement{
		ID: "12",
	}
	out, err := json.Marshal(policy)
	if err != nil {
		t.Errorf("unexpected error; got: %s", err)
	}
	want := "{\"Sid\":\"12\",\"Effect\":\"\"}"
	if string(out) != want {
		t.Errorf("unexpected out; want: %s got %s", want, string(out))
	}
}

func TestIAMPolicyStatementNotEmptyJSON(t *testing.T) {
	policy := IAMPolicyStatement{
		ID:        "12",
		Resource:  StrOrSlice{"iam:"},
		Principal: map[string]StrOrSlice{"AWS": StrOrSlice{"iam"}},
		Condition: map[string]interface{}{"String": "matching ARN"},
	}
	out, err := json.Marshal(policy)
	if err != nil {
		t.Errorf("unexpected error; got: %s", err)
	}
	want := `{"Sid":"12","Effect":"","Resource":"iam:","Principal":{"AWS":"iam"},"Condition":{"String":"matching ARN"}}`
	if string(out) != want {
		t.Errorf("unexpected out; want: %s got %s", want, string(out))
	}
}

func TestIAMPolicy_Marshal(t *testing.T) {

	tests := []struct {
		name    string
		policy  IAMPolicy
		want    []byte
		wantErr bool
	}{
		{
			name: "IAMPolicy Marshal",
			policy: IAMPolicy{
				Version: "12",
				Statement: []IAMPolicyStatement{
					{
						Effect:   "Allow",
						Action:   StrOrSlice{"iam:*"},
						Resource: StrOrSlice{"*"},
					},
				},
			},
			want: []byte(`{"Version":"12","Statement":[{"Effect":"Allow","Action":"iam:*","Resource":"*"}]}`),
		},
		{
			name: "IAMPolicy Not Marshal",
			policy: IAMPolicy{
				Version: "12",
				Statement: []IAMPolicyStatement{
					{
						Effect:      "Allow",
						NotAction:   StrOrSlice{"iam:*"},
						NotResource: StrOrSlice{"*"},
					},
				},
			},
			want: []byte(`{"Version":"12","Statement":[{"Effect":"Allow","NotAction":"iam:*","NotResource":"*"}]}`),
		},
		{
			name: "IAMPolicy Marshal Lists",
			policy: IAMPolicy{
				Version: "12",
				Statement: []IAMPolicyStatement{
					{
						Effect:   "Allow",
						Action:   StrOrSlice{"act:1", "act:2"},
						Resource: StrOrSlice{"res1", "res2"},
					},
				},
			},
			want: []byte(`{"Version":"12","Statement":[{"Effect":"Allow","Action":["act:1","act:2"],"Resource":["res1","res2"]}]}`),
		},
		{
			name: "IAMPolicy Principal",
			policy: IAMPolicy{
				Version: "12",
				Statement: []IAMPolicyStatement{
					{
						Effect: "Allow",
						Action: StrOrSlice{"sts:AssumeRole"},
						Principal: map[string]StrOrSlice{
							"AWS": StrOrSlice{"root"},
						},
					},
				},
			},
			want: []byte(`{"Version":"12","Statement":[{"Effect":"Allow","Action":"sts:AssumeRole","Principal":{"AWS":"root"}}]}`),
		},
		{
			name: "IAMPolicy NotPrincipal",
			policy: IAMPolicy{
				Version: "12",
				Statement: []IAMPolicyStatement{
					{
						Effect: "Deny",
						Action: StrOrSlice{"sts:AssumeRole"},
						NotPrincipal: map[string]StrOrSlice{
							"AWS": StrOrSlice{"root"},
						},
					},
				},
			},
			want: []byte(`{"Version":"12","Statement":[{"Effect":"Deny","Action":"sts:AssumeRole","NotPrincipal":{"AWS":"root"}}]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := json.Marshal(tt.policy); (err != nil) != tt.wantErr {
				t.Errorf("IAMPolicy Marshal() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("IAMPolicy Marshal() = %v, want %v", string(got), string(tt.want))
				}
			}
		})
	}
}
