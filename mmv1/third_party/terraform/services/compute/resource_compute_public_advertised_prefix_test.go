package compute_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-google/google/acctest"
	"github.com/hashicorp/terraform-provider-google/google/envvar"
	"github.com/hashicorp/terraform-provider-google/google/tpgresource"
	transport_tpg "github.com/hashicorp/terraform-provider-google/google/transport"
)

// Since we only have access to one test prefix range we cannot run tests in parallel
func TestAccComputePublicPrefixes(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"delegated_prefix":                     testAccComputePublicDelegatedPrefix_publicDelegatedPrefixesBasicTest,
		"advertised_prefix":                    testAccComputePublicAdvertisedPrefix_publicAdvertisedPrefixesBasicTest,
		"public_delegated_prefixes_ipv6":       testAccComputePublicDelegatedPrefix_publicDelegatedPrefixesIpv6Test,
		"public_advertised_prefixes_pdp_scope": testAccComputePublicAdvertisedPrefix_publicAdvertisedPrefixesPdpScopeTest,
	}

	for name, tc := range testCases {
		// shadow the tc variable into scope so that when
		// the loop continues, if t.Run hasn't executed tc(t)
		// yet, we don't have a race condition
		// see https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccComputePublicAdvertisedPrefix_publicAdvertisedPrefixesPdpScopeTest(t *testing.T) {
	context := map[string]interface{}{
		"description":   envvar.GetTestPublicAdvertisedPrefixDescriptionFromEnv(t),
		"random_suffix": acctest.RandString(t, 10),
	}

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckComputePublicAdvertisedPrefixDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccComputePublicAdvertisedPrefix_publicAdvertisedPrefixesPdpScopeExample(context),
			},
			{
				ResourceName:      "google_compute_public_advertised_prefix.prefix",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccComputePublicAdvertisedPrefix_publicAdvertisedPrefixesPdpScopeExample(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_compute_public_advertised_prefix" "prefix" {
  name = "tf-test-my-prefix%{random_suffix}"
  description = "%{description}"
  dns_verification_ip = "127.127.0.0"
  ip_cidr_range = "127.127.0.0/16"
  pdp_scope = "REGIONAL"
}
`, context)
}

func testAccComputePublicAdvertisedPrefix_publicAdvertisedPrefixesBasicTest(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"description":   envvar.GetTestPublicAdvertisedPrefixDescriptionFromEnv(t),
		"random_suffix": acctest.RandString(t, 10),
	}

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckComputePublicAdvertisedPrefixDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccComputePublicAdvertisedPrefix_publicAdvertisedPrefixesBasicExample(context),
			},
			{
				ResourceName:      "google_compute_public_advertised_prefix.prefix",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccComputePublicAdvertisedPrefix_publicAdvertisedPrefixesBasicExample(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_compute_public_advertised_prefix" "prefix" {
  name = "tf-test-my-prefix%{random_suffix}"
  description = "%{description}"
  dns_verification_ip = "127.127.0.0"
  ip_cidr_range = "127.127.0.0/16"
}
`, context)
}

func testAccComputePublicDelegatedPrefix_publicDelegatedPrefixesBasicTest(t *testing.T) {
	context := map[string]interface{}{
		"description":   envvar.GetTestPublicAdvertisedPrefixDescriptionFromEnv(t),
		"random_suffix": acctest.RandString(t, 10),
	}

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckComputePublicDelegatedPrefixDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccComputePublicDelegatedPrefix_publicDelegatedPrefixesBasicExample(context),
			},
			{
				ResourceName:            "google_compute_public_delegated_prefix.prefix",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"region"},
			},
		},
	})
}

func testAccComputePublicDelegatedPrefix_publicDelegatedPrefixesBasicExample(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_compute_public_advertised_prefix" "advertised" {
  name = "tf-test-my-prefix%{random_suffix}"
  description = "%{description}"
  dns_verification_ip = "127.127.0.0"
  ip_cidr_range = "127.127.0.0/16"
}

resource "google_compute_public_delegated_prefix" "prefix" {
  name = "tf-test-my-prefix%{random_suffix}"
  description = "my description"
  region = "us-central1"
  ip_cidr_range = "127.127.0.0/24"
  parent_prefix = google_compute_public_advertised_prefix.advertised.id
}

resource "google_compute_public_delegated_prefix" "subprefix" {
  name = "tf-test-my-subprefix%{random_suffix}"
  description = "my description"
  region = "us-central1"
  ip_cidr_range = "127.127.0.0/26"
  parent_prefix = google_compute_public_delegated_prefix.prefix.id
}
`, context)
}

func testAccComputePublicDelegatedPrefix_publicDelegatedPrefixesIpv6Test(t *testing.T) {
	context := map[string]interface{}{
		"description":   envvar.GetTestPublicAdvertisedPrefixDescriptionFromEnv(t),
		"random_suffix": acctest.RandString(t, 10),
	}

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckComputePublicDelegatedPrefixDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccComputePublicDelegatedPrefix_publicDelegatedPrefixesIpv6Example(context),
			},
			{
				ResourceName:            "google_compute_public_delegated_prefix.prefix",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"region"},
			},
		},
	})
}

func testAccComputePublicDelegatedPrefix_publicDelegatedPrefixesIpv6Example(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_compute_public_advertised_prefix" "advertised" {
  name = "tf-test-ipv6-pap%{random_suffix}"
  description = "%{description}"
  dns_verification_ip = "2001:db8::"
  ip_cidr_range = "2001:db8::/32"
  pdp_scope = "REGIONAL"
}

resource "google_compute_public_delegated_prefix" "prefix" {
  name = "tf-test-root-pdp%{random_suffix}"
  description = "test-delegation-mode-pdp"
  region = "us-west1"
  ip_cidr_range = "2001:db8::/40"
  parent_prefix = google_compute_public_advertised_prefix.advertised.id
  mode = "DELEGATION"
}

resource "google_compute_public_delegated_prefix" "subprefix" {
  name = "tf-test-sub-pdp%{random_suffix}"
  description = "test-forwarding-rule-mode-pdp"
  region = "us-west1"
  ip_cidr_range = "2001:db8::/48"
  parent_prefix = google_compute_public_delegated_prefix.prefix.id
  allocatable_prefix_length = 64
  mode = "EXTERNAL_IPV6_FORWARDING_RULE_CREATION"
}
`, context)
}

func testAccCheckComputePublicDelegatedPrefixDestroyProducer(t *testing.T) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for name, rs := range s.RootModule().Resources {
			if rs.Type != "google_compute_public_delegated_prefix" {
				continue
			}
			if strings.HasPrefix(name, "data.") {
				continue
			}

			config := acctest.GoogleProviderConfig(t)

			url, err := tpgresource.ReplaceVarsForTest(config, rs, "{{ComputeBasePath}}projects/{{project}}/regions/{{region}}/publicDelegatedPrefixes/{{name}}")
			if err != nil {
				return err
			}

			billingProject := ""

			if config.BillingProject != "" {
				billingProject = config.BillingProject
			}

			_, err = transport_tpg.SendRequest(transport_tpg.SendRequestOptions{
				Config:    config,
				Method:    "GET",
				Project:   billingProject,
				RawURL:    url,
				UserAgent: config.UserAgent,
			})
			if err == nil {
				return fmt.Errorf("ComputePublicDelegatedPrefix still exists at %s", url)
			}
		}

		return nil
	}
}

func testAccCheckComputePublicAdvertisedPrefixDestroyProducer(t *testing.T) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for name, rs := range s.RootModule().Resources {
			if rs.Type != "google_compute_public_advertised_prefix" {
				continue
			}
			if strings.HasPrefix(name, "data.") {
				continue
			}

			config := acctest.GoogleProviderConfig(t)

			url, err := tpgresource.ReplaceVarsForTest(config, rs, "{{ComputeBasePath}}projects/{{project}}/global/publicAdvertisedPrefixes/{{name}}")
			if err != nil {
				return err
			}

			billingProject := ""

			if config.BillingProject != "" {
				billingProject = config.BillingProject
			}

			_, err = transport_tpg.SendRequest(transport_tpg.SendRequestOptions{
				Config:    config,
				Method:    "GET",
				Project:   billingProject,
				RawURL:    url,
				UserAgent: config.UserAgent,
			})
			if err == nil {
				return fmt.Errorf("ComputePublicAdvertisedPrefix still exists at %s", url)
			}
		}

		return nil
	}
}
