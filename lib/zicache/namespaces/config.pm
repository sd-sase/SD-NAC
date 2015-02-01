package zicache::namespaces::config;

use Data::Dumper;

sub new {
  my ($class) = @_;
  my $self = bless {}, $class;

  $self->init();

  return $self;
}

sub build {
  my ($self) = @_;

  tie %tmp_cfg, 'Config::IniFiles', ( -file => $self->{file} );

  $self->{cfg} = \%tmp_cfg;

  return $self->build_child();
}

sub init {
  # abstact
}

sub build_child {
  # abstract
}

1;
